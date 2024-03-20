package eval

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/grafana/grafana/pkg/expr/classic"
)

func extractEvalString(frame *data.Frame) (s string) {
	if frame == nil {
		return "empty frame"
	}

	if frame.Meta == nil || frame.Meta.Custom == nil {
		return
	}

	if evalMatches, ok := frame.Meta.Custom.([]classic.EvalMatch); ok {
		sb := strings.Builder{}

		for i, m := range evalMatches {
			valString := "null"
			if m.Value != nil {
				if *m.Value == float64(int64(*m.Value)) {
					valString = fmt.Sprintf("%d", int64(*m.Value))
				} else {
					valString = fmt.Sprintf("%.3f", *m.Value)
				}
			}
			sb.WriteString(fmt.Sprintf("%s=%s", m.Metric, valString))

			if i < len(evalMatches)-1 {
				sb.WriteString(", ")
			}
		}
		return sb.String()
	}

	if captures, ok := frame.Meta.Custom.([]NumberValueCapture); ok {
		// Sort captures in ascending order of "Var" so we can assert in tests
		sort.Slice(captures, func(i, j int) bool {
			return captures[i].Var < captures[j].Var
		})
		sb := strings.Builder{}
		for _, capture := range captures {
			if capture.Var == frame.RefID {
				valString := "null"
				if capture.Value != nil {
					if *capture.Value == float64(int64(*capture.Value)) {
						valString = fmt.Sprintf("%d", int64(*capture.Value))
					} else {
						valString = fmt.Sprintf("%.3f", *capture.Value)
					}
				}
				sb.WriteString(fmt.Sprintf("%s=%s", capture.Metric, valString))
			}
		}
		return sb.String()
	}
	return ""
}

// extractValues returns the metric name and value for the result expression in the frame.
// For classic conditions the same metric name can have multiple values due to multiple conditions.
// It returns nil if there are no results in the frame.
func extractValues(frame *data.Frame) map[string]NumberValueCapture {
	if frame == nil {
		return nil
	}
	if frame.Meta == nil || frame.Meta.Custom == nil {
		return nil
	}

	if matches, ok := frame.Meta.Custom.([]classic.EvalMatch); ok {
		// Classic evaluations only have a single match but can contain multiple conditions.
		// Conditions have a strict ordering which we can rely on to distinguish between values,
		// in the case of duplicate names.
		v := make(map[string]NumberValueCapture, len(matches))
		for i, match := range matches {
			// In classic conditions we can use the condition position as a suffix to distinguish between duplicate names.
			// We can guarantee determinism as conditions are ordered and this order is preserved when marshaling.
			metric := match.Metric
			if _, ok := v[metric]; ok {
				metric += fmt.Sprintf(" [%d]", i)
			}
			v[metric] = NumberValueCapture{
				Var:    frame.RefID,
				Labels: match.Labels,
				Value:  match.Value,
				Metric: match.Metric,
			}
		}
		return v
	}

	if captures, ok := frame.Meta.Custom.([]NumberValueCapture); ok {
		v := make(map[string]NumberValueCapture, len(captures))
		for _, capture := range captures {
			if capture.Var == frame.RefID {
				v[capture.Metric] = capture
			}
		}
		return v
	}
	return nil
}
