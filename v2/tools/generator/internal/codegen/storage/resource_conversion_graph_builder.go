/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package storage

import (
	"fmt"

	"github.com/rotisserie/eris"
	"golang.org/x/exp/slices"

	"github.com/Azure/azure-service-operator/v2/tools/generator/internal/astmodel"
)

// ResourceConversionGraphBuilder is used to construct a group conversion graph with all the required conversions
// to/from/between storage variants of the packages
type ResourceConversionGraphBuilder struct {
	name       string                       // Name of the resources needing conversions
	hubVersion string                       // Optional override for which version should be the hub
	references astmodel.InternalTypeNameSet // Set of all Type Names that make up this group
	links      astmodel.TypeAssociation     // A collection of links that make up the graph
}

// NewResourceConversionGraphBuilder creates a new builder for a specific resource/type
func NewResourceConversionGraphBuilder(name string) *ResourceConversionGraphBuilder {
	return &ResourceConversionGraphBuilder{
		name:       name,
		references: astmodel.NewInternalTypeNameSet(),
		links:      make(astmodel.TypeAssociation),
	}
}

// SetHubVersion configures which version should be treated as the hub/storage version,
// overriding the default behavior where preview versions link backward (making the oldest preview the hub).
// When set, all preview versions link forward instead, making the newest (specified) version the hub.
func (b *ResourceConversionGraphBuilder) SetHubVersion(hubVersion string) {
	b.hubVersion = hubVersion
}

// Add includes the supplied package reference(s) in the conversion graph for this group
func (b *ResourceConversionGraphBuilder) Add(names ...astmodel.InternalTypeName) {
	for _, name := range names {
		b.references.Add(name)
	}
}

// Build connects all the provided API definitions together into a single conversion graph
func (b *ResourceConversionGraphBuilder) Build() (*ResourceConversionGraph, error) {
	stages := []func([]astmodel.InternalTypeName){
		b.apiReferencesConvertToStorage,
		b.compatibilityReferencesConvertToOriginalPackage,
		b.previewReferencesConvertBackward,
		b.nonPreviewReferencesConvertForward,
	}

	toProcess := make([]astmodel.InternalTypeName, 0, len(b.references))
	for name := range b.references {
		toProcess = append(toProcess, name)
	}

	slices.SortFunc(
		toProcess,
		func(i astmodel.InternalTypeName, j astmodel.InternalTypeName) int {
			return astmodel.ComparePathAndVersion(
				i.PackageReference().ImportPath(),
				j.PackageReference().ImportPath())
		})

	for _, s := range stages {
		s(toProcess)
		toProcess = b.withoutLinkedNames(toProcess)
	}

	if b.hubVersion != "" {
		// When hubVersion is set, we may have separate GA and preview chains (e.g. types that share
		// names across classic ARO and HCP). Each chain has its own hub, so multiple remaining items are valid.
		if len(toProcess) == 0 {
			return nil, eris.Errorf(
				"expected at least one hub reference for %q, but all references are linked",
				b.name)
		}
	} else {
		// Expect to have only the hub reference left
		if len(toProcess) != 1 {
			return nil, eris.Errorf(
				"expected to have linked all references in with name %q, but have %d left",
				b.name,
				len(toProcess))
		}
	}

	result := &ResourceConversionGraph{
		name:  b.name,
		links: b.links,
	}

	return result, nil
}

// compatibilityReferencesConvertToOriginalPackage links any compatibility references to the original if present
func (b *ResourceConversionGraphBuilder) compatibilityReferencesConvertToOriginalPackage(names []astmodel.InternalTypeName) {
	for i, name := range names {
		// Last package can't be linked forward
		if i+1 >= len(names) {
			break
		}

		if !isCompatibilityPackage(name.InternalPackageReference()) {
			continue
		}

		next := names[i+1]
		if next.InternalPackageReference().HasAPIVersion(name.InternalPackageReference().APIVersion()) {
			b.links[name] = next
		}
	}
}

// apiReferencesConvertToStorage links each API type to the associated storage package
func (b *ResourceConversionGraphBuilder) apiReferencesConvertToStorage(names []astmodel.InternalTypeName) {
	for _, name := range names {
		if s, ok := name.InternalPackageReference().(astmodel.DerivedPackageReference); ok {
			n := name.WithPackageReference(s.Base())
			b.links[n] = name
		}
	}
}

// previewReferencesConvertBackward links each preview version to the immediately prior version, no matter whether it's
// preview or GA.
// When a hubVersion override is configured, preview versions are linked forward among themselves only,
// keeping them separate from GA versions to avoid cross-contamination between resources that share
// sub-type names (e.g. classic ARO's IngressProfile vs HCP's IngressProfile).
func (b *ResourceConversionGraphBuilder) previewReferencesConvertBackward(names []astmodel.InternalTypeName) {
	if b.hubVersion != "" {
		// Hub version override: link preview versions forward to each other only.
		var previewNames []astmodel.InternalTypeName
		for _, name := range names {
			if name.InternalPackageReference().IsPreview() {
				previewNames = append(previewNames, name)
			}
		}

		for i, name := range previewNames {
			if i+1 >= len(previewNames) {
				break
			}

			b.links[name] = previewNames[i+1]
		}

		return
	}

	for i, name := range names {
		if i == 0 || !name.InternalPackageReference().IsPreview() {
			continue
		}

		b.links[name] = names[i-1]
	}
}

// nonPreviewReferencesConvertForward links each version with the immediately following version.
// By the time we run this stage, we should only have non-preview (aka GA) releases left.
// When a hubVersion override is configured, only non-preview versions are linked to avoid
// creating cross-links between GA and preview chains.
func (b *ResourceConversionGraphBuilder) nonPreviewReferencesConvertForward(names []astmodel.InternalTypeName) {
	toLink := names
	if b.hubVersion != "" {
		// Only link non-preview versions forward to keep GA and preview chains separate.
		var nonPreview []astmodel.InternalTypeName
		for _, name := range names {
			if !name.InternalPackageReference().IsPreview() {
				nonPreview = append(nonPreview, name)
			}
		}

		toLink = nonPreview
	}

	for i, name := range toLink {
		if i+1 >= len(toLink) {
			break
		}

		b.links[name] = toLink[i+1]
	}
}

// withoutLinkedNames returns a new slice of references, omitting any that are already linked into our graph
func (b *ResourceConversionGraphBuilder) withoutLinkedNames(
	names []astmodel.InternalTypeName,
) []astmodel.InternalTypeName {
	result := make([]astmodel.InternalTypeName, 0, len(names))
	for _, ref := range names {
		if _, ok := b.links[ref]; !ok {
			result = append(result, ref)
		}
	}

	return result
}

func isCompatibilityPackage(ref astmodel.PackageReference) bool {
	switch r := ref.(type) {
	case astmodel.ExternalPackageReference:
		return false
	case astmodel.LocalPackageReference:
		return r.HasVersionPrefix("v1api")
	case astmodel.SubPackageReference:
		return isCompatibilityPackage(r.Parent())
	default:
		msg := fmt.Sprintf(
			"unexpected PackageReference implementation %T",
			ref)
		panic(msg)
	}
}
