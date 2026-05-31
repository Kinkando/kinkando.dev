This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

## Pre Commit Lint

1. Install `husky` and `lint-staged`

```bash
pnpm install --save-dev husky lint-staged prettier eslint
pnpm i -D eslint-config-prettier eslint-plugin-simple-import-sort eslint-plugin-simple-import-sort eslint-plugin-unused-imports typescript-eslint]
```

optional: `git config core.hooksPath frontend/.husky` for support git hook sub-directory

2. Add script `prepare` in `package.json`

```json
{
  "scripts": {
    "format": "prettier --write .",
    "lint": "eslint --fix .",
    "prepare": "husky init"
  },
  "lint-staged": {
    "src/**/*.{js,jsx,ts,tsx,svelte,vue}": ["prettier --write", "eslint --fix", "git add"],
    "*.{json,md,scss}": ["prettier --write", "git add"]
  }
}
```

3. Then run command `pnpm prepare` or `npx husky init` to auto-generated `.husky/*` files
4. Edit file `.husky/pre-commit` like this

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
