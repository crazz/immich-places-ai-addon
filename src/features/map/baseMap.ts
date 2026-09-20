import {MAP_DEFAULT_CENTER, MAP_DEFAULT_ZOOM, MAP_TILE_MAX_ZOOM} from '@/utils/map';

import {TILE_ATTRIBUTION} from './constant';

import type L from 'leaflet';

export function createBaseMap(leaflet: typeof L, container: HTMLElement, tileURL: string): {map: L.Map; tiles: L.TileLayer} {
 const map = leaflet.map(container, {attributionControl: false, zoomControl: false}).setView(MAP_DEFAULT_CENTER, MAP_DEFAULT_ZOOM);
 leaflet.control.attribution({prefix: false}).addTo(map);
 const tiles = leaflet.tileLayer(tileURL, {attribution: TILE_ATTRIBUTION, maxZoom: MAP_TILE_MAX_ZOOM}).addTo(map);
 return {map, tiles};
}
