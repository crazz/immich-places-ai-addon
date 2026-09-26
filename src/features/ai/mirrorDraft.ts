import {isCurrentDescription} from './descriptionWriteTypes';
import {isMirrorSelection} from './mirrorTypes';

import type {TDraft} from './draftTypes';
import type {TMirrorPlan, TMirrorSelection} from './mirrorTypes';

export function isCurrentMirror(selection: TMirrorSelection, draft: TDraft): boolean {
 return isMirrorSelection(selection) && (!selection.direction || !draft.headingStale) && (!selection.precision || !draft.radiusStale) && (!selection.place || !!draft.candidateId) && (selection.languages ?? []).every(language => isCurrentDescription(draft.descriptions.find(item => item.language === language), draft.factsRevision));
}

export function mirrorMatchesDraft(mirror: TMirrorPlan, draft: TDraft): boolean {
 const choice = draft.mirror; const value = mirror.value;
 if (!choice || !isCurrentMirror(choice, draft) || ['direction', 'precision', 'place', 'provenance', 'model'].some(key => choice[key as keyof TMirrorSelection] !== mirror.selection[key as keyof TMirrorSelection]) || [...(choice.languages ?? [])].sort().join(',') !== [...(mirror.selection.languages ?? [])].sort().join(',') || value.review.draftRevision !== draft.revision || value.review.factsRevision !== draft.factsRevision) {return false;}
 if ((choice.languages ?? []).some(language => value.descriptions?.[language] !== draft.descriptions.find(item => item.language === language)?.text)) {return false;}
 const method = draft.heading === null ? 'unknown' : draft.headingUserSupplied ? 'user_supplied' : draft.headingMethod || 'unknown';
 const uncertainty = draft.heading === null || draft.headingUserSupplied || !draft.headingMethod ? null : draft.headingUncertainty ?? null;
 if (value.direction && (value.direction.heading !== draft.heading || value.direction.method !== method || value.direction.uncertainty !== uncertainty)) {return false;}
 if (value.precision && (value.precision.radius !== draft.radius || value.precision.basis !== draft.radiusBasis)) {return false;}
 return !value.provenance || value.provenance.reviewedAt === draft.updatedAt;
}
