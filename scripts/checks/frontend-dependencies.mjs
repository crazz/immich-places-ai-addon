import path from 'node:path';

import ts from 'typescript';


function diagnosticText(diagnostic) {
	return ts.flattenDiagnosticMessageText(diagnostic.messageText, ' ');
}

function staticImports(source) {
	const imports = [];
	function visit(node) {
		if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) && node.moduleSpecifier) {
			imports.push({node, specifier: node.moduleSpecifier.text});
		} else if (ts.isImportEqualsDeclaration(node) && ts.isExternalModuleReference(node.moduleReference)) {
			imports.push({node, specifier: node.moduleReference.expression.text});
		} else if (ts.isImportTypeNode(node) && ts.isLiteralTypeNode(node.argument) && ts.isStringLiteral(node.argument.literal)) {
			imports.push({node, specifier: node.argument.literal.text});
		} else if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'require' && node.arguments.length === 1 && ts.isStringLiteral(node.arguments[0])) {
			imports.push({node, specifier: node.arguments[0].text});
		}
		ts.forEachChild(node, visit);
	}
	visit(source);
	return imports;
}

function runtimeReferences(source, program) {
	const references = new Set();
	if (source.isDeclarationFile) {
		return references;
	}
	const result = program.emit(source, () => {}, undefined, false, {
		after: [
() => node => {
			function visit(current) {
				references.add(ts.getOriginalNode(current));
				ts.forEachChild(current, visit);
			}
			visit(node);
			return node;
		}
]
	});
	if (result.emitSkipped || result.diagnostics.length) {
		throw new Error(`${source.fileName}: TypeScript runtime analysis failed: ${result.diagnostics.map(diagnosticText).join('; ') || 'emit skipped'}`);
	}
	return references;
}

function localAsset(specifier, absolute, options, host, root) {
	if (/\.[cm]?[jt]sx?$/.test(specifier) || !path.extname(specifier)) {
		return undefined;
	}
	const candidates = specifier.startsWith('.') ? [path.resolve(path.dirname(absolute), specifier)] : [];
	for (const [pattern, targets] of Object.entries(options.paths ?? {})) {
		const [prefix, suffix = ''] = pattern.split('*');
		if (pattern === specifier || (pattern.includes('*') && specifier.startsWith(prefix) && specifier.endsWith(suffix))) {
			const matched = specifier.slice(prefix.length, suffix ? -suffix.length : undefined);
			candidates.push(...targets.map(target => path.resolve(options.baseUrl ?? options.pathsBasePath ?? root, target.replace('*', matched))));
		}
	}
	const resolvedFileName = candidates.find(candidate => host.fileExists(candidate));
	return resolvedFileName ? {resolvedFileName} : undefined;
}

function localSpecifier(specifier, options) {
	return specifier.startsWith('.') || specifier.startsWith('/') || Object.keys(options.paths ?? {}).some(pattern => {
		const [prefix, suffix = ''] = pattern.split('*');
		return pattern.includes('*') ? specifier.startsWith(prefix) && specifier.endsWith(suffix) : specifier === pattern;
	});
}

export function frontendGraph({root, paths, host}) {
	const files = paths.filter(name => /\.[cm]?[jt]sx?$/.test(name));
	const config = ts.readConfigFile(path.join(root, 'tsconfig.json'), host.readFile);
	if (config.error) {
		throw new Error(`tsconfig.json: configuration error: ${diagnosticText(config.error)}`);
	}
	const {options, errors} = ts.parseJsonConfigFileContent(config.config, host, root, undefined, path.join(root, 'tsconfig.json'));
	if (errors.length) {
		throw new Error(`tsconfig.json: configuration error: ${errors.map(diagnosticText).join('; ')}`);
	}
	const emitOptions = {...options, allowJs: true, noEmit: false, noEmitOnError: false, declaration: false, emitDeclarationOnly: false, outDir: path.join(root, '.dependency-check-emitted')};
	const compilerHost = {
		...ts.createCompilerHost(emitOptions),
...host,
		useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames,
		getSourceFile(absolute, languageVersion) {
			const content = host.readFile(absolute);
			return content === undefined ? undefined : ts.createSourceFile(absolute, content, languageVersion, true);
		}
	};
	const parser = ts.createProgram({rootNames: files.map(name => path.join(root, name)), options: emitOptions, host: compilerHost});
	const edges = [];
	for (const name of files) {
		const absolute = path.join(root, name);
		const source = parser.getSourceFile(absolute);
		const diagnostics = parser.getSyntacticDiagnostics(source);
		if (diagnostics.length) {
			throw new Error(`${name}: TypeScript parse error: ${diagnostics.map(diagnosticText).join('; ')}`);
		}
		const runtime = runtimeReferences(source, parser);
		for (const {node, specifier} of staticImports(source)) {
			const resolved = ts.resolveModuleName(specifier, absolute, options, host).resolvedModule ?? localAsset(specifier, absolute, options, host, root);
			if (!resolved && localSpecifier(specifier, options)) {
				throw new Error(`${name}: unresolved local import ${specifier}`);
			}
			if (resolved && !resolved.isExternalLibraryImport) {
				edges.push({from: name, to: path.relative(root, resolved.resolvedFileName), runtime: runtime.has(node)});
			}
		}
	}
	return {files, edges};
}

export function frontendViolations({edges}) {
	return edges.filter(({from, to}) => !from.startsWith('src/features/ai/') && to.startsWith('src/features/ai/') && !/^src\/features\/ai\/index\.[cm]?[jt]sx?$/.test(to))
		.map(({from, to}) => `${from} -> ${to}: use the AI public integration surface src/features/ai/index.ts`);
}
