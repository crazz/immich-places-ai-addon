import L from 'leaflet';
import {expect, it} from 'vitest';

import {MAP_DEFAULT_CENTER, MAP_DEFAULT_ZOOM} from '@/utils/map';

import {createBaseMap} from './baseMap';

it('creates only the configured basemap and preserves default view without manual editing actions', () => {
 const container = document.createElement('div'); document.body.append(container);
 const {map, tiles} = createBaseMap(L, container, '/synthetic/{z}/{x}/{y}.png');
 try {
  expect(map.getCenter().lat).toBe(MAP_DEFAULT_CENTER[0]); expect(map.getZoom()).toBe(MAP_DEFAULT_ZOOM);
  expect(map.hasLayer(tiles)).toBe(true); expect(tiles.options.maxZoom).toBe(19);
  let count = 0; map.eachLayer(() => {count++;}); expect(count).toBe(1);
  map.fire('click', {latlng: L.latLng(1, 2)}); map.fire('contextmenu', {latlng: L.latLng(3, 4)});
  count = 0; map.eachLayer(() => {count++;}); expect(count).toBe(1);
 } finally {map.remove(); container.remove();}
});
