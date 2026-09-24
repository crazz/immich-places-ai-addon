'use client';
import L from 'leaflet';
import {useEffect, useRef, useState} from 'react';

import {createBaseMap, useStreetTileURL} from '@/features/map';

import type {TDraftPoint} from './draftTypes';
import type {ReactElement} from 'react';

export function DraftMap({point, onPoint}: {point: TDraftPoint | null; onPoint: (point: TDraftPoint) => void}): ReactElement {
 const container = useRef<HTMLDivElement>(null); const mapRef = useRef<L.Map | null>(null);
 const handler = useRef(onPoint); handler.current = onPoint;
 const [hasError, setHasError] = useState(false); const [hasTileError, setHasTileError] = useState(false);
 const tileURL = useStreetTileURL();
 useEffect(() => {
  if (!container.current) {return;}
  let observer: ResizeObserver | undefined;
  try {
   const {map, tiles} = createBaseMap(L, container.current, tileURL); mapRef.current = map;
   map.setView([0, 0], 2, {animate: false}); L.control.zoom({position: 'topright'}).addTo(map);
   map.on('click', (event: L.LeafletMouseEvent) => {const location = event.latlng.wrap(); if (Math.abs(location.lat) <= 90) {handler.current({latitude: location.lat, longitude: location.lng});}});
   tiles.on('tileerror', () => setHasTileError(true));
   if (typeof ResizeObserver !== 'undefined') {observer = new ResizeObserver(() => map.invalidateSize({pan: false})); observer.observe(container.current);}
  } catch {setHasError(true);}
  return () => {observer?.disconnect(); mapRef.current?.remove(); mapRef.current = null;};
 }, [tileURL]);
 useEffect(() => {
  const map = mapRef.current; if (!map || !point) {return;}
  const node = document.createElement('span'); node.textContent = '■ Draft camera';
  node.className = 'block w-max rounded bg-white p-1 text-black';
  const marker = L.marker([point.latitude, point.longitude], {draggable: true, title: 'Draft camera', icon: L.divIcon({html: node, className: 'ai-draft-camera', iconSize: [24, 24]})}).addTo(map);
  marker.on('dragend', () => {const location = marker.getLatLng().wrap(); if (Math.abs(location.lat) <= 90) {handler.current({latitude: location.lat, longitude: location.lng});}});
  map.setView([point.latitude, point.longitude], Math.max(map.getZoom(), 12), {animate: false});
  return () => {marker.off(); marker.remove();};
 }, [point, tileURL]);
 return <section role={'region'} aria-label={'Editable draft map'} onDrop={event => {event.preventDefault(); event.stopPropagation();}} onDragOver={event => event.preventDefault()}>
  <p>{'Click the map or drag the draft camera. Save explicitly to persist changes.'}</p>
  <div ref={container} className={'h-64 w-full rounded border'} />
  {hasError && <p role={'status'}>{'Draft map unavailable. Numeric editing remains available.'}</p>}
  {hasTileError && <p role={'status'}>{'Draft map tiles unavailable. Numeric editing remains available.'}</p>}
        </section>;
}
