export type TReviewCoordinate = {latitude: number; longitude: number};
export type TReviewCamera = TReviewCoordinate & {granularity: string; radius: number | null; radiusBasis: string};
export type TReviewDirection = {degrees: number; uncertainty: number | null; method: string};
export type TReviewCandidate = {id: string; name: string; locality: string | null; country: string | null; camera: TReviewCamera | null; subject: {name: string; location: TReviewCoordinate | null} | null; direction: TReviewDirection | null; support: string; observations: string[]; sources: string[]; limitations: string[]};
export type TReviewDescription = {language: string; status: 'complete' | 'unavailable'; text: string | null; basis: 'scene_only' | 'candidate'; candidateId: string | null; reason: string | null};
export type TReviewSource = {id: string; kind: string; lineage: string; url?: string; title?: string | null; relevance?: string};
export type TReview = {outcome: string; initialCandidate: string | null; candidates: TReviewCandidate[]; observations: {id: string; kind: string; text: string}[]; sources: TReviewSource[]; omissions: string[]; descriptions: TReviewDescription[]; languages: string[]; primaryLanguage: string; warnings: string[]; mode: string};
