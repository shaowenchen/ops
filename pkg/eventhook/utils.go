package eventhook

import "strings"

// replaceWildcards replaces wildcard "*" in urlTemplate with corresponding values from sourceSubject
// Examples:
//
//	urlTemplate: "ops.clusters.*.namespaces.*.pods.*.alerts"
//	sourceSubject: "ops.clusters.cluster1.namespaces.ns1.pods.pod1.events"
//	result: "ops.clusters.cluster1.namespaces.ns1.pods.pod1.alerts"
//
//	urlTemplate: "ops.clusters.*.nodes.*.alerts"
//	sourceSubject: "ops.clusters.cluster1.nodes.node1.events"
//	result: "ops.clusters.cluster1.nodes.node1.alerts"
func replaceWildcards(urlTemplate, sourceSubject string) string {
	// Split both strings by "."
	templateParts := strings.Split(urlTemplate, ".")
	sourceParts := strings.Split(sourceSubject, ".")

	// Build result by matching template parts with source parts
	result := make([]string, len(templateParts))
	sourceIndex := 0

	for i, templatePart := range templateParts {
		if templatePart == "*" {
			// Replace wildcard with corresponding value from source
			if sourceIndex < len(sourceParts) {
				result[i] = sourceParts[sourceIndex]
				sourceIndex++
			} else {
				// If source is shorter, keep the wildcard (shouldn't happen in normal cases)
				result[i] = "*"
			}
		} else {
			// Keep the template part as-is (e.g., "ops", "clusters", "namespaces", "pods", "alerts")
			result[i] = templatePart
			// If this part matches the source at current index, advance source index
			// This handles cases where template has fixed parts that should match source
			if sourceIndex < len(sourceParts) {
				if templatePart == sourceParts[sourceIndex] {
					// Matches, advance both
					sourceIndex++
				}
				// If doesn't match, we still keep the template part and don't advance source
				// This allows template to have different final parts (e.g., "alerts" vs "events")
			}
		}
	}

	return strings.Join(result, ".")
}
