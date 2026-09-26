import type {TDraft, TDraftPoint} from './draftTypes';

export type TDraftEdit = {mirror?: TDraft['mirror'] | null; reviewPrecision?: boolean; camera?: TDraftPoint | null; fields?: TDraft['fields']; primaryLanguage?: string; descriptionPolicy?: TDraft['descriptionPolicy']; heading?: number | null; reviewHeading?: boolean; descriptions?: Record<string, {text?: string; review?: boolean}>; state?: TDraft['state']; candidateId?: string};
export function cameraInput(latitude: string, longitude: string): TDraftPoint | null {
 if (!latitude.trim() && !longitude.trim()) {return null;}
 const lat = Number(latitude); const lon = Number(longitude);
 if (!latitude.trim() || !longitude.trim() || !Number.isFinite(lat) || !Number.isFinite(lon) || Math.abs(lat) > 90 || Math.abs(lon) > 180) {throw new Error('Enter both camera coordinates within latitude −90…90 and longitude −180…180.');}
 return {latitude: lat, longitude: lon};
}
