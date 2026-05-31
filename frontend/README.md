# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```bash
# create a new project in the current directory
npx sv create

# create a new project in my-app
npx sv create my-app
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```bash
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```bash
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## Pre Commit Lint

1. Install `husky` and `lint-staged`

```bash
pnpm install --save-dev husky lint-staged prettier eslint
pnpm i -D eslint-config-prettier eslint-plugin-simple-import-sort eslint-plugin-simple-import-sort eslint-plugin-unused-imports typescript-eslint
```

2. Add script `prepare` in `package.json`

```json
{
  "scripts": {
    "format": "prettier --write .",
    "lint": "eslint --fix .",
    "prepare": "husky install"
  },
  "lint-staged": {
    "src/**/*.{js,jsx,ts,tsx,svelte,vue}": ["prettier --write", "eslint --fix", "git add"],
    "*.{json,md,scss}": ["prettier --write", "git add"]
  }
}
```

3. Then run command `pnpm prepare` to auto-generated `.husky/_/*` files
4. Edit file `.husky/_/pre-commit` like this

```bash
#!/bin/sh

pnpm lint-staged --no-stash
```

5. Edit eslint configuration on file `eslint.config.js`

```js
const prettier = require('eslint-config-prettier');
const js = require('@eslint/js');
const globals = require('globals');
const ts = require('typescript-eslint');
const simpleImportSort = require('eslint-plugin-simple-import-sort');
const unusedImports = require('eslint-plugin-unused-imports');

module.exports = ts.config(
  js.configs.recommended,
  ...ts.configs.recommended,
  prettier,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      },
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: __dirname
      }
    }
  },
  {
    ignores: ['build/', 'dist/', '.svelte-kit/']
  },
  {
    plugins: {
      'simple-import-sort': simpleImportSort,
      'unused-imports': unusedImports
    },
    rules: {
      '@typescript-eslint/no-require-imports': 'off',
      '@typescript-eslint/no-unused-vars': 'warn',
      '@typescript-eslint/member-ordering': 'warn',
      '@typescript-eslint/consistent-type-definitions': 'warn',
      '@typescript-eslint/no-magic-numbers': 'warn',
      '@typescript-eslint/consistent-type-imports': 'warn',
      '@typescript-eslint/no-unnecessary-condition': 'warn',
      '@typescript-eslint/explicit-member-accessibility': 'warn',
      '@typescript-eslint/typedef': 'warn',
      'simple-import-sort/imports': 'warn',
      'simple-import-sort/exports': 'warn',
      'unused-imports/no-unused-imports': 'error',
      'unused-imports/no-unused-vars': [
        'warn',
        {
          vars: 'all',
          varsIgnorePattern: '^_',
          args: 'after-used',
          argsIgnorePattern: '^_'
        }
      ]
    }
  }
);
```

6. Edit prettier configuration on file `.prettierrc`

```json
{
  "useTabs": false,
  "singleQuote": true,
  "trailingComma": "none",
  "printWidth": 150
}
```

7. Check your `tsconfig.json` for config your `eslint.config.js` wheter to use ESModule (require) or TSModule (import).
