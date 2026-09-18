import js from '@eslint/js';
import {FlatCompat} from '@eslint/eslintrc';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const compat = new FlatCompat({
	baseDirectory: __dirname,
	recommendedConfig: js.configs.recommended,
	allConfig: js.configs.all
});

const applicationConfig = compat.config({
		root: true,
		env: {
			browser: true,
			node: true,
			es2022: true
		},
		parser: '@typescript-eslint/parser',
		parserOptions: {
			ecmaFeatures: {jsx: true},
			ecmaVersion: 2022,
			sourceType: 'module',
			tsconfigRootDir: __dirname,
			project: ['./tsconfig.json']
		},
		plugins: ['@typescript-eslint', 'react', 'import', 'react-hooks', 'unused-imports', 'eslint-plugin-tsdoc'],
		extends: [
			'eslint:recommended',
			'plugin:@typescript-eslint/recommended',
			'plugin:@next/next/recommended-legacy',
			'plugin:react-hooks/recommended',
			'plugin:import/recommended',
			'plugin:import/typescript',
			'prettier'
		],
		ignorePatterns: [
			'.eslintrc.js',
			'.next',
			'eslint.config.mjs',
			'node_modules',
			'next-env.d.ts',
			'postcss.config.js',
			'postcss.config.mjs',
			'tailwind.config.ts'
		],
		settings: {
			react: {version: 'detect'},
			next: {rootDir: './'},
			'import/resolver': {
				typescript: {
					project: './tsconfig.json'
				}
			}
		},
		rules: {
			'tsdoc/syntax': 'warn',
			'import/default': 0,
			'react/prop-types': 0,
			'no-async-promise-executor': 0,
			quotes: [2, 'single', {avoidEscape: true}],
			'object-curly-spacing': [2, 'never'],
			'array-bracket-spacing': [2, 'never'],
			semi: 'error',
			'no-else-return': ['error', {allowElseIf: false}],
			'eol-last': ['error', 'always'],
			'import/no-named-as-default-member': 2,
			'array-bracket-newline': ['error', {multiline: true}],
			'react/jsx-curly-brace-presence': ['error', {props: 'always', children: 'always'}],
			'react/jsx-first-prop-new-line': ['error', 'multiline'],
			'react/jsx-closing-tag-location': 2,
			'no-unused-vars': 'error',
			'unused-imports/no-unused-imports': 'error',
			'unused-imports/no-unused-vars': [
				'error',
				{
					vars: 'all',
					varsIgnorePattern: '^_',
					args: 'after-used',
					argsIgnorePattern: '^_'
				}
			],
			'import/first': 2,
			'import/newline-after-import': 2,
			'import/no-duplicates': 2,
			curly: ['error', 'all'],
			'object-curly-newline': [
				'error',
				{
					ObjectExpression: {multiline: true, consistent: true},
					ObjectPattern: {multiline: true, consistent: true},
					ImportDeclaration: {multiline: true, consistent: true},
					ExportDeclaration: {multiline: true, minProperties: 3}
				}
			],
			'object-property-newline': ['error', {allowAllPropertiesOnSameLine: true}],
			'@typescript-eslint/consistent-type-imports': [
				2,
				{
					prefer: 'type-imports',
					disallowTypeAnnotations: true,
					fixStyle: 'separate-type-imports'
				}
			],
			'@typescript-eslint/no-var-requires': 0,
			'@typescript-eslint/no-unused-vars': ['error', {argsIgnorePattern: '^_'}],
			'@typescript-eslint/no-explicit-any': [1],
			'@typescript-eslint/array-type': ['error', {default: 'array'}],
			'@typescript-eslint/consistent-type-assertions': 0,
			'@typescript-eslint/consistent-type-definitions': ['error', 'type'],
			'@typescript-eslint/consistent-indexed-object-style': ['error', 'record'],
			'@typescript-eslint/explicit-function-return-type': [
				'error',
				{
					allowExpressions: true,
					allowTypedFunctionExpressions: true,
					allowHigherOrderFunctions: false,
					allowDirectConstAssertionInArrowFunctions: false,
					allowConciseArrowFunctionExpressionsStartingWithVoid: false,
					allowedNames: []
				}
			],
			'@typescript-eslint/naming-convention': [
				'error',
				{
					selector: 'default',
					format: ['camelCase', 'PascalCase'],
					filter: {
						regex: '^(@typescript-eslint/|react/|react-hooks/|no-|import/|object-|array-|eol-|tailwindcss/|unused-imports/|sort-|newlines-between|brace-style|comma-dangle).*$',
						match: false
					}
				},
				{selector: 'function', format: ['camelCase', 'PascalCase']},
				{
					selector: 'variableLike',
					format: ['camelCase', 'PascalCase', 'UPPER_CASE'],
					leadingUnderscore: 'allow'
				},
				{
					selector: 'variable',
					types: ['boolean'],
					format: ['PascalCase'],
					prefix: ['is', 'are', 'should', 'has', 'can', 'did', 'will', 'with', 'was', 'only', 'have']
				},
				{
					selector: 'variableLike',
					format: ['PascalCase'],
					filter: {regex: '(Context)$|(ContextApp)$|^Component$', match: true}
				},
				{selector: ['typeParameter', 'typeAlias'], format: ['PascalCase'], prefix: ['T']},
				{selector: 'interface', format: ['PascalCase'], prefix: ['I']}
			],
			'@typescript-eslint/no-misused-promises': ['error', {checksConditionals: true, checksVoidReturn: false}],
			'@typescript-eslint/no-non-null-asserted-nullish-coalescing': 'error',
			'@typescript-eslint/no-unnecessary-qualifier': 'error',
			'@typescript-eslint/no-unnecessary-type-arguments': 'error',
			'@typescript-eslint/no-unnecessary-boolean-literal-compare': [
				'error',
				{
					allowComparingNullableBooleansToTrue: false,
					allowComparingNullableBooleansToFalse: false
				}
			],
			'@typescript-eslint/prefer-for-of': 'error',
			'@typescript-eslint/prefer-function-type': 'error',
			'@typescript-eslint/prefer-includes': 'error',
			'@typescript-eslint/promise-function-async': 'error',
			'@typescript-eslint/require-array-sort-compare': 'error',
			'brace-style': 'off',
			'comma-dangle': 'off',
			'@typescript-eslint/prefer-optional-chain': 'error',
			indent: 'off',
			'@typescript-eslint/indent': 0,
			'no-multi-spaces': ['error', {ignoreEOLComments: false}],
			'no-mixed-spaces-and-tabs': ['warn', 'smart-tabs'],
			'react/jsx-max-props-per-line': 'off',
			'react-hooks/exhaustive-deps': [
				'warn',
				{
					additionalHooks: '(^useAsyncTrigger$|^useDeepCompareMemo$)'
				}
			],
			'import/no-unresolved': 0, //Issue with package exports
			'sort-imports': [
				'error',
				{
					ignoreCase: false,
					ignoreDeclarationSort: true,
					ignoreMemberSort: false,
					memberSyntaxSortOrder: ['none', 'all', 'multiple', 'single'],
					allowSeparatedGroups: true
				}
			],
			'import/order': [
				'error',
				{
					groups: [
						'builtin', // Built-in imports (come from NodeJS native) go first
						'external', // <- External imports
						'internal', // <- Absolute imports
						['sibling', 'parent'], // <- Relative imports, the sibling and parent types they can be mingled together
						'index', // <- index imports
						'unknown', // <- unknown
						'type' // <- Types go last
					],
					pathGroups: [
						{
							pattern: '@/**',
							group: 'internal'
						}
					],
					'newlines-between': 'always',
					alphabetize: {
						/* sort in ascending order. Options: ["ignore", "asc", "desc"] */
						order: 'asc',
						/* ignore case. Options: [true, false] */
						caseInsensitive: true
					},
					warnOnUnassignedImports: true,
					pathGroupsExcludedImportTypes: ['type']
				}
			],
			'import/consistent-type-specifier-style': ['error', 'prefer-top-level']
		}
});

const toolingRules = Object.fromEntries(applicationConfig.flatMap(config =>
	Object.entries(config.rules ?? {}).filter(([name]) => !name.includes('/') || name.startsWith('import/'))
));

export default [
	...applicationConfig.map(config => Object.keys(config).length === 1 && config.ignores
		? config
		: {...config, ignores: [...(config.ignores ?? []), 'scripts/checks/**/*.mjs']}),
	{
		files: ['scripts/checks/**/*.mjs'],
		languageOptions: {
			ecmaVersion: 2022,
			sourceType: 'module',
			globals: {Buffer: 'readonly', console: 'readonly', process: 'readonly'}
		},
		plugins: {import: applicationConfig.find(config => config.plugins?.import).plugins.import},
		rules: {...js.configs.recommended.rules, ...toolingRules}
	}
];
