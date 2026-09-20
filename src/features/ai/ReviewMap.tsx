'use client';
import L from 'leaflet';
import {useEffect, useMemo, useRef, useState} from 'react';

import {createBaseMap, useStreetTileURL} from '@/features/map';

import {reviewGeometry} from './reviewGeometry';
import {reviewMapLayers} from './reviewMapLayers';

import type {TReview} from './reviewTypes';
import type {ReactElement} from 'react';

export function ReviewMap({review, focus}: {review: TReview; focus: string | null}): ReactElement {
 const container = useRef<HTMLDivElement>(null); const mapRef = useRef<L.Map | null>(null); const tilesRef = useRef<L.TileLayer | null>(null);
 const streetTileURL = useStreetTileURL();
 const [isTileUnavailable, setIsTileUnavailable] = useState(false); const [hasMapError, setHasMapError] = useState(false);
 const [frame, setFrame] = useState(0);
 const geometry = useMemo(() => reviewGeometry(review, focus), [review, focus]);
 const candidate = review.candidates.find(item => item.id === focus) || null;
 const hasGeometry = geometry.markers.length > 0;
 useEffect(() => {
  if (!container.current || !hasGeometry) {return;}
  let observer: ResizeObserver | undefined;
  try {
   const {map, tiles} = createBaseMap(L, container.current, streetTileURL);
   mapRef.current = map; tilesRef.current = tiles;
   L.control.zoom({position: 'topright'}).addTo(map);
   tiles.on('tileerror', () => setIsTileUnavailable(true));
   if (typeof ResizeObserver !== 'undefined') {observer = new ResizeObserver(() => map.invalidateSize({pan: false})); observer.observe(container.current);}
  } catch {setHasMapError(true);}
  return () => {observer?.disconnect(); tilesRef.current?.off(); mapRef.current?.remove(); mapRef.current = null; tilesRef.current = null;};
 }, [streetTileURL, hasGeometry]);
 useEffect(() => {
  const map = mapRef.current; if (!map || !hasGeometry) {return;}
  const layers = reviewMapLayers(L, geometry, candidate);
  try {
   layers.addTo(map);
   map.fitBounds(geometry.bounds.map(point => [point.latitude, point.longitude]), {padding: [32, 32], maxZoom: 14, animate: false});
  } catch {setHasMapError(true);}
  return () => {layers.remove(); layers.clearLayers();};
 }, [geometry, candidate, streetTileURL, hasGeometry, frame]);
 return <section role={'region'} aria-label={'Read-only proposal map'} className={'space-y-2'} onDragOver={event => event.preventDefault()} onDrop={event => {event.preventDefault(); event.stopPropagation();}} onContextMenu={event => {event.preventDefault(); event.stopPropagation();}}>
  <p className={'text-xs'}>{'■ Camera · ◆ Subject · Alternative markers are proposals, not approved locations.'}</p>
  {hasGeometry ? <div ref={container} className={'h-80 w-full rounded border border-(--color-border)'} /> : <p>{'No supported map geometry. Original numeric information remains available below.'}</p>}
  {hasGeometry && <button type={'button'} onClick={() => setFrame(value => value + 1)}>{'Frame proposal'}</button>}
  {(geometry.degraded || (candidate?.camera?.radius ?? 0) > 20_000_000) && <p role={'status'}>{'Some geometry exceeds map display limits. Original numeric coordinates and radius remain authoritative.'}</p>}
  {hasMapError && <p role={'status'}>{'Map unavailable. Use the complete numeric and text presentation below.'}</p>}
  {isTileUnavailable && <><p role={'status'}>{'Map tiles unavailable. All coordinates and evidence remain available below.'}</p><button type={'button'} onClick={() => {setIsTileUnavailable(false); tilesRef.current?.redraw();}}>{'Retry map tiles'}</button></>}
  <p className={'text-xs'}>{'Dashed circle: uncalibrated radius estimate. Arrow: estimated true-north heading, with no distance or field-of-view claim.'}</p>
        </section>;
}
