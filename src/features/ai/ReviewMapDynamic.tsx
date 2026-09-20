import {createMapDynamic} from '@/features/map';

export const ReviewMapDynamic = createMapDynamic(async () => (await import('./ReviewMap')).ReviewMap);
