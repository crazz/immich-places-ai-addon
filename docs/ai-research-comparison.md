# Research comparison

Run `node scripts/ai/research-comparison.mjs /absolute/path/input.json > out/research-comparison.json` from the repository root. Use only approved photos and keep private inputs/reports outside Git. This command reads a local report input; it does not contact a provider or upload images. Run its tests with `node --test scripts/ai/research-comparison.test.mjs`.

Each case binds one photo SHA-256 and the exact hint. Record baseline and Research runs against those same inputs. Optional album/date context must also be identical; record it in the hint/context description used for the comparison. Keep failed, unknown, ambiguous and coarse outcomes. A report is evidence of recorded results, not proof that the model performed live web research.

```json
[
  {
    "id": "approved-photo-1",
    "photoSHA256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "hint": "Exact approved hint, with identical optional context",
    "reference": {"latitude": 50, "longitude": 14, "source": "Independently established camera position", "uncertaintyM": 20},
    "runs": [
      {
        "label": "research",
        "provider": "configured profile name and revision",
        "model": "configured model",
        "promptVersion": "research-v1",
        "schemaVersion": "2.0",
        "status": "located",
        "coordinates": {"latitude": 50.001, "longitude": 14.002},
        "estimatedErrorM": 500,
        "latencyMs": 240000,
        "links": ["https://example.org/reference"]
      }
    ]
  }
]
```

Add a baseline entry to `runs` with its actual prompt/schema versions. Omit `reference` when the camera position is not independently known. For unknown or failed runs use `coordinates: null`, `estimatedErrorM: null`, and retain a safe failure category such as `timeout`. For ambiguous runs retain the alternatives as an additional `alternatives` array; use `coordinates: null` unless the model supplied a preferred candidate. A run may repeat `photoSHA256` and `hint` to assert matching inputs; disagreement is rejected.

`estimatedErrorM` is the model's uncertainty estimate. `actualErrorM` is the great-circle distance to the optional independent camera reference, using a mean Earth radius of 6,371,008.8 meters. The report retains the independent reference uncertainty as `referenceUncertaintyM`. Without a reference or its known finite, nonnegative `uncertaintyM`, measured error is `null` and `measurement` is `not_measured`; proposed coordinates remain available. The distance is relative to that reference, not a guaranteed absolute accuracy. Neither number creates an acceptance cutoff. The user decides whether a proposal is useful.

The output records configuration, prompt/schema, coordinates, both error measures, latency, links and failures. `inputDigest` binds the matched photo/hint without echoing the case's private hint. Reported links are ordinary answer content; this command does not fetch or certify them.
