import {expect, it} from 'vitest';

import {translationRun} from './testing/translation';
import {isTranslationPage, isTranslationRun} from './translationTypes';

it('validates bounded private translation history without accepting inconsistent language outcomes', () => {
 const run = translationRun();
 expect(isTranslationRun(run)).toBe(true);
 expect(isTranslationPage({ids: [run.id], next: ''})).toBe(true);
 for (const value of [null, {...run, items: [...run.items, run.items[0]]}, {...run, items: [{...run.items[0], language: 'fr'}]}, {...run, items: [{...run.items[0], text: null}, run.items[1]]}, {...run, request: {...run.request, basis: 'x'.repeat(16385)}}, {...run, request: {...run.request, confirmed: false}}]) {expect(isTranslationRun(value)).toBe(false);}
 for (const value of [null, {ids: ['foreign'], next: ''}, {ids: Array(21).fill(run.id), next: ''}, {ids: [], next: 1}]) {expect(isTranslationPage(value)).toBe(false);}
});
