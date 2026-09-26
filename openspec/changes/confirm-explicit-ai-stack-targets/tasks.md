## 1. Review and freeze an exact stack GPS selection

- [x] 1.1 Deliver explicit per-member source/GPS observation and bounded exact selection with the analyzed-photo default, valid selected GPS, independent authorization and no implicit stack expansion (S01–S04, S20; design D1).
- [x] 1.2 Deliver immutable per-target field/baseline comparisons and membership handling that preserve analyzed-photo descriptions and reject changed or unreadable selected targets (S05–S07, P01–P10; design D2).
- [x] 1.3 Deliver keyboard-accessible full-manifest confirmation and manual-overlap resolution for every selected GPS target without altering other pending work (S17–S18, P14–P16, D05–D07; design D5).

## 2. Confirm and execute independently guarded targets

- [x] 2.1 Deliver atomic all-target durable approval and cross-version/account exclusion with private-safe overlap rejection and exact per-target fields (S08, W01–W11; design D2–D3).
- [x] 2.2 Deliver independent per-target attempts, generations, sender completion and explicit retry/reconciliation that preserve successful siblings through restart and lost acknowledgements (S09–S10, W14–W18, W27–W28; design D3).

## 3. Retain honest partial outcomes as conditions change

- [x] 3.1 Deliver per-target expiry/conflict/unresolved outcomes and revision protection across edits, disablement, shutdown, owner deletion and installation changes (S11–S13, R01–R02, W23–W25; design D4).
- [x] 3.2 Deliver independent verified catalog refresh and durable accessible partial history, preserving sync fences, missing-row behavior and private response ordering (S14, S19, W19–W22; design D5).

## 4. Upgrade without expanding earlier approvals

- [x] 4.1 Preserve old single-photo bytes, attempt counts, audit and unresolved guards through the additive upgrade and reopen, including overlap with new multi-target plans on real SQLite (S15; design D3–D4).
- [x] 4.2 Preserve single-photo, description, manual/GPX and disabled-AI behavior while requiring separate live stack evidence, with [verification-plan.md](verification-plan.md) and applicable engineering gates satisfied (S16, W26; design D5).
