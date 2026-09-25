# Local AI review drafts

AI Results supports explicit local acceptance of retained valid Visual, Context-assisted and Research proposals. Acceptance creates one durable private draft per analysis. Accepting again retrieves the current draft; it never replaces corrections. Failed, canceled and corrupt results remain separate from valid proposals.

## Review workflow

1. Open a retained result in AI Results and choose **Accept as local draft**. Ambiguous proposals require an explicit candidate. Unknown and subject-only proposals keep camera coordinates absent.
2. Choose a candidate, enter a complete camera pair, or click/drag the separate draft map. Numeric controls work without tiles or images. Zero is valid; clearing both camera fields records an absent point. The original proposal remains visible above the editor.
3. Save local corrections. Camera/candidate changes mark inherited heading, radius and location-dependent descriptions for renewed review. Scene-only text retains its independent basis. Each language can be corrected or reviewed independently. Heading is optional and accepts zero inclusive to 360 exclusive.
4. Select GPS for the analyzed photo and stage the draft. Coarse estimates and stale or unavailable non-GPS fields do not block staging. Staging grants no Immich-write permission. Changing a staged draft, including acknowledging a new baseline, produces an unstaged revision; stage that revision separately.
5. Reject or reopen a draft without deleting the original analysis or another run's decision.

AI editing does not call providers, modify Immich or populate manual/GPX pending locations. An overlapping manual pending location blocks AI staging until that manual choice is explicitly saved or discarded. Both decisions remain available.

Every save uses the current revision. A conflicting or unacknowledged save preserves the editor and blocks replay until **Compare saved revision** retrieves current state. Choose to discard local edits or retain them for another explicit save. Ordinary navigation offers keep/discard; a page exit warns about unsaved edits. Account/result changes abort and fence old private responses.

## Current source and GPS baseline

Local acceptance and edits work while source metadata is offline or unavailable. New drafts have an unavailable baseline; that is different from an observed photo with no GPS.

**Review current source and GPS** explicitly reads currently authorized metadata, displays the current photo and exact nullable GPS, and creates a five-minute observation bound to the draft revision. A loaded image plus the explicit review checkbox enables acknowledgement. Acknowledgement rereads metadata and access before saving a new revision. Changed GPS/source/access, expired observations or another saved revision require renewed review. Failed review never clears a previous saved baseline.

Original analysis provenance remains unchanged. The reviewed image identity is separately versioned from asset ID, upstream owner, type and checksum; metadata timestamps remain part of the original source-digest contract. A changed original fingerprint is displayed without claiming that the old analysis saw the current image. Observation storage is bounded to ten unexpired observations per draft, with at most 100 expired rows removed per cleanup pass.

## Storage and recovery

Migration 026 adds owner/installation-qualified drafts, immutable revision snapshots and expiring baseline observations. Draft references protect source history from ordinary job cleanup. Catalog resets and disabled provider execution preserve local decisions. Account deletion cascades only that account's private resources; installation rotation does not transfer authority.

Back up SQLite consistently before deployment. Roll back application binaries with draft tables retained. Do not run destructive down migrations as recovery: they remove private decisions. A failed migration rolls back transactionally. Retained tables remain compatible with earlier analysis-read queries; live deployment/rollback has not been exercised.

CH18 now adds [exact read-only previews](ai-write-previews.md) for reviewed staged GPS revisions. CH19's confirmed GPS writer remains planned. No live provider/Immich compatibility is claimed by synthetic tests. See [draft implementation verification](../openspec/changes/persist-revisioned-ai-review-drafts/implementation-verification.md).
