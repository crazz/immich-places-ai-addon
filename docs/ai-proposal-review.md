# Read-only AI proposal inspection

Open **AI Results**, then a saved result. Inspection shows the retained proposal independently of whether the current source photo remains accessible. [History](ai-results-history.md) owns authorization, immutable run identity and current-image access.

Camera and subject points have different shapes and explicit labels. Every candidate retains its own numeric coordinates, granularity, limitations and optional radius. Zero is a valid coordinate, radius or heading. Missing values remain unknown. Radius is a supplied, uncalibrated estimate with its original basis; it is not a confidence interval or proof of exact accuracy.

A located result initially focuses its canonical candidate. An ambiguous result starts with all alternatives. **Inspect candidate** and **Show all alternatives** change only the local map focus. They do not approve a candidate, create a draft or alter a saved result.

Heading is the stored angle clockwise from true north, with its method and nullable uncertainty. It is never calculated from the subject position. An arrow is shown only with a camera point and supplied heading whose uncertainty is known and below 90 degrees. Other valid direction estimates remain available numerically. The arrow does not assert range, pitch, roll or field of view.

The map shares the existing street-tile configuration and base Leaflet setup. It owns isolated, non-draggable layers with no manual edit actions. Initial or candidate focus frames the geometry; tile retries and language changes do not reset normal pan/zoom. Antimeridian framing uses the shortest longitude span. Points outside the Web Mercator latitude range and circles larger than 20,000 km have an explicit display-limit notice; original numeric values remain authoritative. Tile or image failure leaves the complete text presentation usable.

**Evidence and limitations** distinguishes model observations from retained Context provenance. User hints and unknown lineage are not independent corroboration. Only retained source identifiers, kinds and lineage are shown; private source content and public citations are not reconstructed. Provider text is inert text, including marker labels. Opening evidence performs no research or provider call.

Description tabs show exactly the requested language set and mark the primary language. Each tab retains its complete text or unavailable reason, and its scene-only or candidate-specific basis. Focusing another candidate never reassigns an existing description. Arrow keys, Home and End navigate language tabs. Candidate buttons and evidence disclosure work from the keyboard, and all meaningful spatial information has a numeric/text equivalent.

Closing or changing a result removes its map layers and local inspection state. Account changes also clear private detail and abort stale reads through the history boundary. Existing manual/GPX state remains in its original composition. Inspection cannot change pending coordinates, call manual save, accept a draft or write Immich. CH16 owns revisioned draft editing; CH17 owns later translation operations.

No new dependency, migration, provider request or deployment configuration is introduced. See [verification evidence](engineering/ai-proposal-review-verification.md) for synthetic tests and their limits.
