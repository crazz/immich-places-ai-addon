import type {TReviewCandidate} from './reviewTypes';
import type {ReactElement} from 'react';

export function CandidateFacts({candidate, isResearch = false}: {candidate: TReviewCandidate; isResearch?: boolean}): ReactElement {
 const {camera, subject, direction} = candidate;
 const estimateFormat: Intl.NumberFormatOptions = {maximumSignificantDigits: 3, notation: (camera?.radius ?? 0) >= 1e9 ? 'scientific' : 'standard'};
 return <div className={'space-y-1 text-sm'}>
  <p>{camera ? `Camera: ${camera.latitude}, ${camera.longitude}` : 'Camera: Unknown'}</p>
  {isResearch && <p className={'font-semibold'}>{camera?.radius == null ? 'Estimated error: unknown' : `Estimated error: ±${camera.radius >= 1000 ? `${(camera.radius / 1000).toLocaleString('en', estimateFormat)} km` : `${camera.radius.toLocaleString('en', estimateFormat)} m`}`}</p>}
  {isResearch && <p className={'whitespace-pre-wrap'}>{candidate.support}</p>}
  <p>{subject?.location ? `Subject: ${subject.location.latitude}, ${subject.location.longitude}` : 'Subject location: Unknown'}</p>
  {subject && <p>{`Subject name: ${subject.name}`}</p>}
  <p>{`Granularity: ${camera?.granularity || 'unknown'}`}</p>
  {!isResearch && <p>{camera?.radius != null ? `Estimated radius: ${camera.radius} m · ${camera.radiusBasis.replaceAll('_', ' ')} · uncalibrated` : 'Precision: unknown'}</p>}
  {camera?.radius != null && <p className={'text-xs text-(--color-text-secondary)'}>{'The supplied estimate is not a confidence interval or proof of exact accuracy.'}</p>}
  <p>{direction ? `Heading: ${direction.degrees}° clockwise from true north` : 'Heading: Unknown'}</p>
  {direction && <p>{`Uncertainty: ${direction.uncertainty === null ? 'Unknown' : `${direction.uncertainty}°`} · Method: ${direction.method.replaceAll('_', ' ')}`}</p>}
  {direction && (direction.uncertainty === null || direction.uncertainty >= 90) && <p className={'text-xs'}>{'Direction is too uncertain for a map arrow; the original numeric estimate is shown above.'}</p>}
  <p className={'text-xs text-(--color-text-secondary)'}>{'Subject position does not establish the camera position or optical-axis heading.'}</p>
  {candidate.limitations.length > 0 && <ul className={'list-disc pl-5'}>{candidate.limitations.map((text, index) => <li key={index}>{text}</li>)}</ul>}
        </div>;
}
