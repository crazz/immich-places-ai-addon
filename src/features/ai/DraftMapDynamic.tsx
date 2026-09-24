import {createMapDynamic} from '@/features/map';

export const DraftMapDynamic = createMapDynamic(async () => (await import('./DraftMap')).DraftMap);
