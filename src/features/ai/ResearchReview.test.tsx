import {cleanup, render, screen} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {ProposalReview} from './ProposalReview';
import {isResultDetail} from './resultDetailTypes';
import {parseReview} from './reviewParser';
import {canonicalProposal, reviewDetail} from './testing/review';

import type {TResultDetail} from './resultDetailTypes';

type TResearchFixture = Omit<TResultDetail, 'proposal'> & {
	proposal: Omit<ReturnType<typeof canonicalProposal>, 'schema_version'> & Record<'schema_version', '2.0'>;
};

vi.mock('./ReviewMapDynamic', () => ({ReviewMapDynamic: () => <p>{'Map unavailable'}</p>}));

function researchDetail(radius: number | null = 5000): TResearchFixture {
	const detail = reviewDetail();
	const proposal = canonicalProposal();
	const camera = proposal.candidates[0].camera_location;
	Object.assign(
		camera,
		{granularity: 'city'},
		Object.fromEntries([
			['estimated_radius_m', radius],
			['radius_basis', radius === null ? 'unknown' : 'model_estimate']
		])
	);
	return {
		...detail,
		entry: {...detail.entry, mode: 'research'},
		proposal: Object.assign(
			proposal,
			Object.fromEntries([['schema_version', '2.0']])
		) as TResearchFixture['proposal'],
		provenance: {
			...detail.provenance,
			Mode: 'research',
			SchemaVersion: '2.0',
			ValidationVersion: 'analysis-result-v2',
			PromptVersion: 'research-v1',
			Context: {Version: 'context-v1', Sources: [], Omissions: []}
		}
	};
}

it('shows Research coordinates and estimated error even with a degraded map', () => {
	const detail = researchDetail();
	expect(isResultDetail(detail)).toBe(true);
	if (!isResultDetail(detail)) {
		throw new Error('Research detail rejected');
	}
	const review = parseReview(detail);
	expect(review?.candidates[0].camera).toMatchObject({
		latitude: 50,
		longitude: 14,
		radius: 5000,
		granularity: 'city'
	});
	if (!review) {
		throw new Error('Research proposal rejected');
	}
	render(<ProposalReview review={review} />);
	expect(screen.getByText('Camera: 50, 14')).toBeVisible();
	expect(screen.getByText('Estimated error: ±5 km')).toBeVisible();
	expect(screen.getByText('Map unavailable')).toBeVisible();
	expect(screen.queryByRole('button', {name: /accept|save|apply/i})).not.toBeInTheDocument();
});

it('shows answer-provided references as deliberate links beside the proposed coordinates', () => {
	const base = researchDetail(500);
	const detail = {
		...base,
		proposal: {
			...base.proposal,
			sources: [
				{
					id: 'reference',
					url: 'https://example.org/photo',
					title: 'Reference photo',
					relevance: 'The church facade matches.'
				}
			],
			candidates: [{...base.proposal.candidates[0], ...Object.fromEntries([['source_refs', ['reference']]])}]
		}
	};
	if (!isResultDetail(detail)) {
		throw new Error('Research detail rejected');
	}
	const review = parseReview(detail);
	expect(review).not.toBeNull();
	if (!review) {
		throw new Error('Answer sources rejected');
	}
	render(<ProposalReview review={review} />);
	const link = screen.getByRole('link', {name: 'Reference photo'});
	expect(link).toHaveAttribute('href', 'https://example.org/photo');
	expect(link).toHaveAttribute('rel', 'noopener noreferrer');
	expect(screen.getByText('The church facade matches.')).toBeVisible();
	expect(screen.getByText('Estimated error: ±500 m')).toBeVisible();
	expect(screen.queryByRole('img')).not.toBeInTheDocument();
});

it('keeps coordinates with missing or unsafe references and renders source text inertly', () => {
	for (const url of [
		'',
		'javascript:alert(1)',
		'https://user:secret@example.org/photo',
		'http://127.0.0.1/a',
		'http://[::1]/a',
		'http://nas.local/a',
		'https://example.org/%0aheader'
	]) {
		const base = researchDetail(null);
		const detail = {
			...base,
			proposal: {
				...base.proposal,
				sources: [
					{
						id: 'reference',
						url,
						title: '<img src=x onerror=alert(1)>',
						relevance: '<script>reference</script>'
					}
				],
				candidates: [{...base.proposal.candidates[0], ...Object.fromEntries([['source_refs', ['reference', 'missing']]])}]
			}
		};
		if (!isResultDetail(detail)) {
			throw new Error('Research detail rejected');
		}
		const review = parseReview(detail);
		expect(review).not.toBeNull();
		if (!review) {
			throw new Error('Unusable reference discarded coordinates');
		}
		render(<ProposalReview review={review} />);
		expect(screen.getByText('Camera: 50, 14')).toBeVisible();
		expect(screen.getByText('Estimated error: unknown')).toBeVisible();
		expect(screen.queryByRole('link')).not.toBeInTheDocument();
		expect(screen.queryByRole('img')).not.toBeInTheDocument();
		expect(screen.getByText('<script>reference</script>')).toBeVisible();
		expect(document.querySelector('script')).toBeNull();
		cleanup();
	}
});

it('rejects malformed Research source collections before rendering', () => {
	const base = researchDetail();
	expect(isResultDetail({...base, proposal: {...base.proposal, sources: {url: 'https://example.org'}}})).toBe(false);
});

it('keeps coordinates and links when an answer source reuses an input source ID', () => {
	const base = researchDetail();
	const detail = {
		...base,
		provenance: {
			...base.provenance,
			Context: {Version: 'context-v1', Sources: [{ID: 'hint', Kind: 'user_hint', Lineage: ''}], Omissions: []}
		},
		proposal: {
			...base.proposal,
			sources: [{id: 'hint', url: 'https://example.org/reference', title: 'Reference', relevance: 'Matching facade'}],
			candidates: [{...base.proposal.candidates[0], ...Object.fromEntries([['source_refs', ['hint']]])}]
		}
	};
	if (!isResultDetail(detail)) {
		throw new Error('Invalid fixture');
	}
	const review = parseReview(detail);
	expect(review).not.toBeNull();
	if (!review) {
		throw new Error('Source collision discarded coordinates');
	}
	render(<ProposalReview review={review} />);
	expect(screen.getByText('Camera: 50, 14')).toBeVisible();
	expect(screen.getByRole('link', {name: 'Reference'})).toHaveAttribute('href', 'https://example.org/reference');
});

it('keeps enormous estimated errors compact and numeric coordinates visible', () => {
	const detail = researchDetail(1e100);
	if (!isResultDetail(detail)) {
		throw new Error('Invalid Research detail');
	}
	const review = parseReview(detail);
	if (!review) {
		throw new Error('Large estimate rejected');
	}
	render(<ProposalReview review={review} />);
	expect(screen.getByText('Camera: 50, 14')).toBeVisible();
	expect(screen.getByText('Estimated error: ±1E97 km')).toBeVisible();
});

it('shows optional answer links even when the candidate omits a local source reference', () => {
	const base = researchDetail();
	const detail = {
		...base,
		proposal: {
			...base.proposal,
			sources: [
				{
					id: 'unreferenced',
					url: 'https://example.org/useful',
					title: 'Useful reference',
					relevance: 'A public comparison photo.'
				}
			]
		}
	};
	if (!isResultDetail(detail)) {
		throw new Error('Invalid Research detail');
	}
	const review = parseReview(detail);
	if (!review) {
		throw new Error('Result rejected');
	}
	render(<ProposalReview review={review} />);
	expect(screen.getByRole('link', {name: 'Useful reference'})).toHaveAttribute('href', 'https://example.org/useful');
});
