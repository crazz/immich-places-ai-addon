import {useEffect, useState} from 'react';

import {AIRequestError} from './aiRequest';

type TReadState<T> = {key: string; data: T | null; loading: boolean; error: string};
export function useResultRead<T>(owner: string, requestKey: string, load: (signal: AbortSignal) => Promise<T>): Omit<TReadState<T>, 'key'> & {refresh: () => void} {
 const key = `${owner}:${requestKey}`;
 const [state, setState] = useState<TReadState<T>>({key, data: null, loading: true, error: ''});
 const [generation, setGeneration] = useState(0);
 useEffect(() => {
  const controller = new AbortController();
  setState(current => ({key, data: current.key === key ? current.data : null, loading: !!owner, error: ''}));
  if (owner) {
   void load(controller.signal).then(data => {
    if (!controller.signal.aborted) {setState({key, data, loading: false, error: ''});}
   }).catch(error => {
    if (controller.signal.aborted) {return;}
    const isDenied = error instanceof AIRequestError && ['UNAUTHENTICATED', 'RESULT_UNAVAILABLE'].includes(error.code);
    setState(current => ({key, data: isDenied ? null : current.data, loading: false, error: isDenied ? 'Result unavailable.' : 'Could not refresh. Displayed information may be stale.'}));
   });
  }
  return () => controller.abort();
 }, [owner, key, load, generation]);
 const visible = state.key === key ? state : {data: null, loading: !!owner, error: ''};
 return {...visible, refresh: () => setGeneration(value => value + 1)};
}
