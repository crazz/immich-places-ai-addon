import type {TReview, TReviewCoordinate} from './reviewTypes';

export type TReviewMarker = {id: string; candidateId: string; kind: 'camera' | 'subject'; label: string; point: TReviewCoordinate};
export type TReviewGeometry = {markers: TReviewMarker[]; bounds: TReviewCoordinate[]; degraded: boolean};

export function reviewGeometry(review: TReview, focus: string | null): TReviewGeometry {
 const candidates = focus ? review.candidates.filter(item => item.id === focus) : review.candidates;
 const markers: TReviewMarker[] = [];
 let isDegraded = false;
 for (const candidate of candidates) {
  for (const kind of ['camera', 'subject'] as const) {
   const point = kind === 'camera' ? candidate.camera : candidate.subject?.location;
   if (!point) {continue;}
   if (Math.abs(point.latitude) > 85.05112878) {isDegraded = true; continue;}
   markers.push({id: `${candidate.id}-${kind}`, candidateId: candidate.id, kind, label: `${focus ? '' : 'Alternative · '}${kind === 'camera' ? 'Camera' : 'Subject'} · ${candidate.name}`, point: {latitude: point.latitude, longitude: point.longitude}});
  }
 }
 const longitudes = markers.map(item => (item.point.longitude + 360) % 360).sort((a, b) => a - b);
 let start = longitudes[0] ?? 0; let largestGap = -1;
 for (let i = 0; i < longitudes.length; i++) {
  const next = longitudes[(i + 1) % longitudes.length];
  const gap = (i + 1 === longitudes.length ? next + 360 : next) - longitudes[i];
  if (gap > largestGap) {largestGap = gap; start = next;}
 }
 const bounds = markers.map(item => {
  let longitude = (item.point.longitude + 360) % 360;
  if (longitude < start) {longitude += 360;}
  if (start > 180) {longitude -= 360;}
  return {latitude: item.point.latitude, longitude};
 });
 return {markers, bounds, degraded: isDegraded};
}
