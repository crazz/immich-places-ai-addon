import type {TReviewGeometry} from './reviewGeometry';
import type {TReviewCandidate} from './reviewTypes';
import type L from 'leaflet';

function markerLabel(label: string, kind: string): HTMLSpanElement {
 const node = document.createElement('span');
 node.textContent = `${kind === 'camera' ? '■' : '◆'} ${label}`;
 node.className = 'block w-max max-w-48 rounded border border-slate-500 bg-white px-2 py-1 text-xs text-slate-900 shadow';
 node.setAttribute('role', 'img'); node.setAttribute('aria-label', label);
 return node;
}
export function reviewMapLayers(leaflet: typeof L, geometry: TReviewGeometry, candidate: TReviewCandidate | null): L.LayerGroup {
 const group = leaflet.layerGroup();
 geometry.markers.forEach((marker, index) => {
  const point = geometry.bounds[index];
  leaflet.marker([point.latitude, point.longitude], {draggable: false, interactive: false, keyboard: false, title: marker.label, icon: leaflet.divIcon({html: markerLabel(marker.label, marker.kind), className: 'ai-review-marker', iconSize: [24, 24], iconAnchor: [12, 12]})}).addTo(group);
 });
 const cameraIndex = geometry.markers.findIndex(marker => marker.kind === 'camera' && marker.candidateId === candidate?.id);
 if (cameraIndex < 0 || !candidate?.camera) {return group;}
 const point = geometry.bounds[cameraIndex]; const center: L.LatLngTuple = [point.latitude, point.longitude];
 if (candidate.camera.radius !== null && candidate.camera.radius <= 20_000_000) {
  leaflet.circle(center, {radius: candidate.camera.radius, color: '#475569', dashArray: '5 5', weight: 1, fillOpacity: 0.08, interactive: false}).addTo(group);
 }
 const direction = candidate.direction;
 if (direction && direction.uncertainty !== null && direction.uncertainty < 90) {
  const label = `Estimated heading ${direction.degrees}° true north; uncertainty ${direction.uncertainty}°`;
  const arrow = document.createElement('span'); arrow.textContent = '↑'; arrow.title = label;
  arrow.setAttribute('role', 'img'); arrow.setAttribute('aria-label', label);
  arrow.style.cssText = `display:block;font-size:44px;line-height:52px;text-align:center;width:52px;height:52px;border:1px dashed #475569;border-radius:50%;color:#334155;opacity:0.7;transform:rotate(${direction.degrees}deg)`;
  leaflet.marker(center, {draggable: false, interactive: false, keyboard: false, title: label, icon: leaflet.divIcon({html: arrow, className: 'ai-review-direction', iconSize: [52, 52], iconAnchor: [26, 52]})}).addTo(group);
 }
 return group;
}
