import {useCallback, useEffect, useRef, useState} from 'react';

import {useBackend} from '@/shared/context/AppContext';

import type {ReactElement} from 'react';

export function VerifiedGPSRefresh({revision}: {revision: number}): ReactElement | null {
 const {refreshDataAction, retryBackendAction} = useBackend();
 const previous = useRef(0); const [hasFailed, setFailed] = useState(false);
 const refresh = useCallback(async (): Promise<void> => {
  try {await Promise.all([refreshDataAction(), retryBackendAction()]); setFailed(false);}
  catch {setFailed(true);}
 }, [refreshDataAction, retryBackendAction]);
 useEffect(() => {
  if (revision > previous.current) {previous.current = revision; void refresh();}
 }, [revision, refresh]);
 return hasFailed ? <div role={'alert'} className={'fixed bottom-2 left-2 z-[1100] rounded border bg-(--color-surface) p-2'}><p>{'GPS verified. Catalog display refresh unavailable.'}</p><button type={'button'} onClick={() => void refresh()}>{'Refresh verified GPS display'}</button></div> : null;
}
