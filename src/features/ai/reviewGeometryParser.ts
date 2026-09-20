import {isRecord} from '@/utils/typeGuards';

import type {TReviewCamera, TReviewCandidate, TReviewCoordinate, TReviewDirection} from './reviewTypes';

export function reviewStrings(value: unknown): value is string[] {return Array.isArray(value) && value.every(item => typeof item === 'string');}
export function nullableText(value: unknown): value is string | null {return value === null || typeof value === 'string';}
export function finiteRange(value: unknown, min: number, max: number): value is number {return typeof value === 'number' && Number.isFinite(value) && value >= min && value <= max;}
export function readCoordinate(value: unknown): TReviewCoordinate | null {
 if (!isRecord(value) || !finiteRange(value.latitude, -90, 90) || !finiteRange(value.longitude, -180, 180)) {return null;}
 return {latitude: value.latitude, longitude: value.longitude};
}
function readCamera(value: unknown, isResearch: boolean): TReviewCamera | null {
 const coordinate = readCoordinate(value);
 if (!coordinate || !isRecord(value) || !['point', 'area', 'site', 'city', 'region'].includes(String(value.granularity)) || !['visual_estimate', 'context_extent', 'source_reported', 'unknown', ...(isResearch ? ['model_estimate'] : [])].includes(String(value.radius_basis)) || (value.estimated_radius_m !== null && !finiteRange(value.estimated_radius_m, 0, Number.MAX_VALUE))) {return null;}
 if ((value.estimated_radius_m === null) !== (value.radius_basis === 'unknown')) {return null;}
 return {...coordinate, granularity: String(value.granularity), radius: value.estimated_radius_m, radiusBasis: String(value.radius_basis)};
}
function readDirection(value: unknown): TReviewDirection | null {
 if (!isRecord(value) || !finiteRange(value.azimuth_deg, 0, 360) || value.azimuth_deg === 360 || value.reference !== 'true_north' || !['visual_estimate', 'known_viewpoint_alignment'].includes(String(value.method)) || (value.uncertainty_deg !== null && !finiteRange(value.uncertainty_deg, 0, 180))) {return null;}
 return {degrees: value.azimuth_deg, uncertainty: value.uncertainty_deg, method: String(value.method)};
}
export function readCandidate(value: unknown, isResearch = false): TReviewCandidate | null {
 if (!isRecord(value) || typeof value.id !== 'string' || !value.id || typeof value.place_name !== 'string' || !nullableText(value.locality) || !nullableText(value.country_code) || typeof value.support_summary !== 'string' || !reviewStrings(value.evidence_refs) || !reviewStrings(value.source_refs) || !reviewStrings(value.uncertainty_notes)) {return null;}
 const camera = readCamera(value.camera_location, isResearch); const direction = readDirection(value.camera_direction);
 if ((value.camera_location !== null && !camera) || (value.camera_direction !== null && (!direction || !camera))) {return null;}
 let subject: TReviewCandidate['subject'] = null;
 if (value.subject !== null) {
  if (!isRecord(value.subject) || typeof value.subject.name !== 'string') {return null;}
  const location = readCoordinate(value.subject.location);
  if (value.subject.location !== null && !location) {return null;}
  subject = {name: value.subject.name, location};
 }
 return {id: value.id, name: value.place_name, locality: value.locality, country: value.country_code, camera, direction, subject, support: value.support_summary, observations: value.evidence_refs, sources: value.source_refs, limitations: value.uncertainty_notes};
}
