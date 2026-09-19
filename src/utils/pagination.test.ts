import {expect, test} from 'vitest';

import {buildPageRange} from './pagination';

test('shows every page when the collection fits the page controls', () => {
	expect(buildPageRange(4, 7)).toEqual([1, 2, 3, 4, 5, 6, 7]);
});

test('shows the current page and its neighbors for a large collection', () => {
	expect(buildPageRange(5, 12)).toEqual([4, 5, 6]);
});

test('keeps the page window inside the available first and last pages', () => {
	expect(buildPageRange(1, 12)).toEqual([1, 2]);
	expect(buildPageRange(12, 12)).toEqual([11, 12]);
});

test('has no page buttons for an empty collection', () => {
	expect(buildPageRange(1, 0)).toEqual([]);
});
