import {useEffect} from 'react';

import {useSelection} from '@/shared/context/AppContext';

export function ManualDraftOverlap({onChange}: {onChange: (ids: string[]) => void}): null {
 const {pendingLocationsByAssetID} = useSelection();
 useEffect(() => {
  onChange(Object.keys(pendingLocationsByAssetID).filter(id => !pendingLocationsByAssetID[id]?.isAlreadyApplied));
  return () => onChange([]);
 }, [pendingLocationsByAssetID, onChange]);
 return null;
}
