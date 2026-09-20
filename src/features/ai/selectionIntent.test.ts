import {expect, it} from 'vitest';

import {selectionIntent} from './selectionIntent';

it('separates selected, current-page and all-matching intentions and blocks unsupported filters', () => {
 const scope = {view: 'all' as const, gpsFilter: 'all', hiddenFilter: 'visible'};
 const input = {scope, selected: ['selected'], page: ['page-a', 'page-b'], blockedReason: ''};
 expect(selectionIntent('selected', input)).toEqual({mode: 'explicit', scope, assetIDs: ['selected']});
 expect(selectionIntent('page', input)).toEqual({mode: 'explicit', scope, assetIDs: ['page-a', 'page-b']});
 expect(selectionIntent('matching', input)).toEqual({mode: 'all-matching', scope});
 expect(() => selectionIntent('matching', {...input, blockedReason: 'GPX status cannot be used for AI selection.'})).toThrow('GPX status');
 expect(() => selectionIntent('selected', {...input, selected: []})).toThrow('Select');
});
