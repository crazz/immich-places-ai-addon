package results

type evidenceSupport struct{ visual, contextExtent, sourceReported, alignment bool }

func geometryFindings(document Document, ctx validationContext) []Finding {
	findings := []Finding{}
	visual := map[string]bool{}
	for _, observation := range document.Observations {
		visual[observation.ID] = observation.Kind == "visual"
	}
	for i, candidate := range document.Candidates {
		support := evidenceSupport{}
		for _, ref := range candidate.EvidenceRefs {
			support.visual = support.visual || visual[ref]
		}
		for _, ref := range candidate.SourceRefs {
			source := ctx.sources[ref]
			support.contextExtent = support.contextExtent || source.ContextExtent
			support.sourceReported = support.sourceReported || source.SourceReportedRadius
			support.alignment = support.alignment || source.ViewpointAlignment
		}
		if direction := candidate.CameraDirection; direction != nil {
			supported := direction.Method == "visual_estimate" && support.visual || direction.Method == "known_viewpoint_alignment" && support.alignment
			if candidate.CameraLocation == nil || !supported {
				findings = append(findings, Finding{Code: "direction_evidence", Path: itemPath("candidates", i, "camera_direction")})
			}
		}

		if location := candidate.CameraLocation; location != nil {
			path := itemPath("candidates", i, "camera_location")
			if location.EstimatedRadiusM == nil {
				if location.RadiusBasis != "unknown" {
					findings = append(findings, Finding{Code: "radius_basis", Path: path + "/radius_basis"})
				}
			} else {
				radius, err := location.EstimatedRadiusM.Float64()
				legacyPrecision := ctx.mode != Research && (location.Granularity == "city" || location.Granularity == "region" || location.Granularity != "point" && radius == 0)
				if err != nil || legacyPrecision {
					findings = append(findings, Finding{Code: "radius_granularity", Path: path + "/estimated_radius_m"})
				}
				supported := location.RadiusBasis == "visual_estimate" && support.visual || location.RadiusBasis == "context_extent" && support.contextExtent || location.RadiusBasis == "source_reported" && support.sourceReported
				supported = supported || ctx.mode == Research && location.RadiusBasis == "model_estimate"
				if !supported {
					findings = append(findings, Finding{Code: "radius_evidence", Path: path + "/radius_basis"})
				}
			}
		}
	}
	return findings
}
