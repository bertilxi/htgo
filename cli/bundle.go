package cli

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/bertilxi/htgo"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

// [Yaffle/TextEncoderTextDecoder.js](https://gist.github.com/Yaffle/5458286)
const textEncoderPolyfill = `function TextEncoder(){}function TextDecoder(){}TextEncoder.prototype.encode=function(e){for(var o=[],t=e.length,r=0;r<t;){var n=e.codePointAt(r),c=0,f=0;for(n<=127?(c=0,f=0):n<=2047?(c=6,f=192):n<=65535?(c=12,f=224):n<=2097151&&(c=18,f=240),o.push(f|n>>c),c-=6;c>=0;)o.push(128|n>>c&63),c-=6;r+=n>=65536?2:1}return o},TextDecoder.prototype.decode=function(e){for(var o="",t=0;t<e.length;){var r=e[t],n=0,c=0;if(r<=127?(n=0,c=255&r):r<=223?(n=1,c=31&r):r<=239?(n=2,c=15&r):r<=244&&(n=3,c=7&r),e.length-t-n>0)for(var f=0;f<n;)c=c<<6|63&(r=e[t+f+1]),f+=1;else c=65533,n=e.length-t;o+=String.fromCodePoint(c),t+=n+1}return o};`

const serverEntry = `import React from "react";
import { renderToString } from "react-dom/server.edge";
import Page from "./$page";

globalThis.renderPage = function renderPage(props) {
  return renderToString(<Page {...props} />);
}`

const clientEntry = `import React from 'react';
import ReactDOM from 'react-dom/client';
import Page from './$page';

const root = ReactDOM.hydrateRoot(
    document.getElementById('page'),
    <Page {...(window.PAGE_PROPS || {})} />
);`

type bundler struct {
	page *htgo.Page
}

func (b *bundler) backendOptions() esbuild.BuildOptions {
	pageDir := path.Dir(b.page.File)
	pageName := path.Base(b.page.File)
	outfile := strings.TrimSuffix(path.Join(htgo.CacheDir, b.page.File), filepath.Ext(b.page.File)) + ".ssr.js"

	return esbuild.BuildOptions{
		Outfile: outfile,
		Stdin: &esbuild.StdinOptions{
			ResolveDir: pageDir,
			Loader:     esbuild.LoaderTSX,
			Contents:   strings.ReplaceAll(serverEntry, "$page", pageName),
		},
		Format:   esbuild.FormatESModule,
		Platform: esbuild.PlatformBrowser,
		Target:   esbuild.ES2020,
		Banner: map[string]string{
			"js": textEncoderPolyfill,
		},
		Loader: map[string]esbuild.Loader{
			".tsx": esbuild.LoaderTSX,
			".css": esbuild.LoaderEmpty,
		},
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  htgo.IsProd(),
		MinifyIdentifiers: htgo.IsProd(),
		MinifySyntax:      htgo.IsProd(),
	}
}

func (b *bundler) buildBackend() (string, error) {
	result := esbuild.Build(b.backendOptions())

	if result.Errors != nil {
		return "", fmt.Errorf("failed to build server bundle: %v", result.Errors)
	}

	return string(result.OutputFiles[0].Contents), nil
}

func (b *bundler) clientOptions() esbuild.BuildOptions {
	pageDir := path.Dir(b.page.File)
	pageName := path.Base(b.page.File)
	outfile := strings.TrimSuffix(path.Join(htgo.CacheDir, b.page.File), filepath.Ext(b.page.File)) + ".js"

	return esbuild.BuildOptions{
		Outfile: outfile,
		Stdin: &esbuild.StdinOptions{
			ResolveDir: pageDir,
			Loader:     esbuild.LoaderTSX,
			Contents:   strings.ReplaceAll(clientEntry, "$page", pageName),
		},
		Format:   esbuild.FormatESModule,
		Platform: esbuild.PlatformBrowser,
		Target:   esbuild.ES2020,
		Loader: map[string]esbuild.Loader{
			".tsx": esbuild.LoaderTSX,
			".css": esbuild.LoaderCSS,
		},
		Plugins: []esbuild.Plugin{
			newTailwindPlugin(htgo.IsProd()),
		},
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  htgo.IsProd(),
		MinifyIdentifiers: htgo.IsProd(),
		MinifySyntax:      htgo.IsProd(),
	}
}

func (b *bundler) buildClient() (string, string, error) {
	result := esbuild.Build(b.clientOptions())

	if result.Errors != nil {
		return "", "", fmt.Errorf("failed to build client bundle: %v", result.Errors)
	}

	jsResult := ""
	cssResult := ""

	for _, file := range result.OutputFiles {
		if strings.HasSuffix(file.Path, ".js") {
			jsResult = string(file.Contents)
		}

		if strings.HasSuffix(file.Path, ".css") {
			cssResult = string(file.Contents)
		}
	}

	return jsResult, cssResult, nil
}

func (b *bundler) watchServer() error {
	ctx, err := esbuild.Context(b.backendOptions())
	if err != nil {
		return err
	}

	err2 := ctx.Watch(esbuild.WatchOptions{})
	if err2 != nil {
		return err2
	}

	return nil
}

func (b *bundler) watchClient() error {
	ctx, err := esbuild.Context(b.clientOptions())
	if err != nil {
		return err
	}

	err2 := ctx.Watch(esbuild.WatchOptions{})
	if err2 != nil {
		return err2
	}

	return nil
}

func (b *bundler) watch() error {
	err := b.watchServer()
	if err != nil {
		return err
	}

	err = b.watchClient()
	if err != nil {
		return err
	}

	return nil
}
