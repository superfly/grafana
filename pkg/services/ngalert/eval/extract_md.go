package eval

import (
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/expr/classic"
)

func extractEvalString(frame *data.Frame, refID string) (s string) {
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
			sb.WriteString(fmt.Sprintf("'%s'=%s", m.Metric, valString))

			if i < len(evalMatches)-1 {
				sb.WriteString(", ")
			}
		}
		return sb.String()
	}

	if caps, ok := frame.Meta.Custom.([]NumberValueCapture); ok {
		sb := strings.Builder{}
		for _, c := range caps {
			if c.Var == refID {
				valString := "null"
				if c.Value != nil {
					if *c.Value == float64(int64(*c.Value)) {
						valString = fmt.Sprintf("%d", int64(*c.Value))
					} else {
						valString = fmt.Sprintf("%.3f", *c.Value)
					}
				}
				sb.WriteString(fmt.Sprintf("'%s'=%s", c.Metric, valString))
			}
		}
		return sb.String()
	}

	return ""
}

// extractValues returns the RefID and value for all classic conditions, reduce, and math expressions in the frame.
// For classic conditions the same refID can have multiple values due to multiple conditions, for them we use the index of
// the condition in addition to the refID to distinguish between different values.
// It returns nil if there are no results in the frame.
func extractValues(frame *data.Frame) map[string]NumberValueCapture {
	if frame == nil {
		return nil
	}
	if frame.Meta == nil || frame.Meta.Custom == nil {
		return nil
	}

	if matches, ok := frame.Meta.Custom.([]classic.EvalMatch); ok {
		// Classic evaluations only have a single match but it can contain multiple conditions.
		// Conditions have a strict ordering which we can rely on to distinguish between values.
		v := make(map[string]NumberValueCapture, len(matches))
		for i, match := range matches {
			// In classic conditions we use refID and the condition position as a way to distinguish between values.
			// We can guarantee determinism as conditions are ordered and this order is preserved when marshaling.
			refID := fmt.Sprintf("%s%d", frame.RefID, i)
			v[refID] = NumberValueCapture{
				Var:    frame.RefID,
				Labels: match.Labels,
				Value:  match.Value,
				Metric: match.Metric,
			}
		}
		return v
	}

	if caps, ok := frame.Meta.Custom.([]NumberValueCapture); ok {
		v := make(map[string]NumberValueCapture, len(caps))
		for _, c := range caps {
			v[c.Var] = c
		}
		return v
	}
	return nil
}
