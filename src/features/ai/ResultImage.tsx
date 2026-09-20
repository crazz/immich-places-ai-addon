import Image from 'next/image';
import {useEffect, useState} from 'react';

import {fetchResultThumbnail} from './resultImageApi';

import type {TResultEntry} from './resultTypes';
import type {ReactElement} from 'react';

export function ResultImage({owner, entry}: {owner: string; entry: TResultEntry}): ReactElement {
 const key = `${owner}:${entry.jobId}:${entry.id}:${entry.sourceAvailable}`;
 const [image, setImage] = useState({key: '', url: ''});
 useEffect(() => {
  const controller = new AbortController(); let url = '';
  setImage({key, url: ''});
  if (owner && entry.sourceAvailable) {
   void fetchResultThumbnail(entry.jobId, entry.id, controller.signal).then(blob => {
    if (controller.signal.aborted) {return;}
    url = URL.createObjectURL(blob); setImage({key, url});
   }).catch(() => {if (!controller.signal.aborted) {setImage({key, url: ''});}});
  }
  return () => {controller.abort(); if (url) {URL.revokeObjectURL(url);}};
 }, [owner, key, entry.jobId, entry.id, entry.sourceAvailable]);
 return image.key === key && image.url ? <Image unoptimized src={image.url} width={320} height={200} alt={'Current source photo'} className={'max-h-48 w-full rounded object-contain'} /> : <p className={'rounded bg-(--color-bg) p-3 text-xs text-(--color-text-secondary)'}>{'Current image unavailable'}</p>;
}
