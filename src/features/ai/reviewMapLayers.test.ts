import L from 'leaflet';
import {expect, it} from 'vitest';

import {reviewGeometry} from './reviewGeometry';
import {reviewMapLayers} from './reviewMapLayers';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

it('creates distinct inert non-draggable points, an estimated circle and only supported direction symbols', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 const candidate = review.candidates[0]; candidate.name = '<img src="https://invented.invalid">'; candidate.direction!.degrees = 0;
 const geometry = reviewGeometry(review, candidate.id);
 const layers = reviewMapLayers(L, geometry, candidate).getLayers();
 expect(layers).toHaveLength(4);
 const markers = layers.filter((item): item is L.Marker => item instanceof L.Marker);
 expect(markers.every(marker => !(marker.options.draggable ?? true))).toBe(true);
 expect(markers.map(marker => marker.options.title)).toContain(`Camera · ${candidate.name}`);
 expect(markers.map(marker => marker.options.title)).toContain(`Subject · ${candidate.name}`);
 const icons = markers.map(marker => marker.getIcon().createIcon());
 expect(icons.some(icon => icon.textContent?.includes('<img'))).toBe(true); expect(icons.every(icon => !icon.querySelector('img'))).toBe(true);
 expect((layers.find(item => item instanceof L.Circle) as L.Circle).getRadius()).toBe(250);
 candidate.direction!.uncertainty = null; candidate.camera!.radius = null;
 expect(reviewMapLayers(L, geometry, candidate).getLayers()).toHaveLength(2);
 candidate.direction!.uncertainty = 120;
 expect(reviewMapLayers(L, geometry, candidate).getLayers()).toHaveLength(2);
});
