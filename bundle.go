package htgo

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

// [Yaffle/TextEncoderTextDecoder.js](https://gist.github.com/Yaffle/5458286)
const textEncoderPolyfill = `function TextEncoder(){}function TextDecoder(){}TextEncoder.prototype.encode=function(e){for(var o=[],t=e.length,r=0;r<t;){var n=e.codePointAt(r),c=0,f=0;for(n<=127?(c=0,f=0):n<=2047?(c=6,f=192):n<=65535?(c=12,f=224):n<=2097151&&(c=18,f=240),o.push(f|n>>c),c-=6;c>=0;)o.push(128|n>>c&63),c-=6;r+=n>=65536?2:1}return o},TextDecoder.prototype.decode=function(e){for(var o="",t=0;t<e.length;){var r=e[t],n=0,c=0;if(r<=127?(n=0,c=255&r):r<=223?(n=1,c=31&r):r<=239?(n=2,c=15&r):r<=244&&(n=3,c=7&r),e.length-t-n>0)for(var f=0;f<n;)c=c<<6|63&(r=e[t+f+1]),f+=1;else c=65533,n=e.length-t;o+=String.fromCodePoint(c),t+=n+1}return o};`

const HtmlTemplate = `<!DOCTYPE html>
<html lang="{{.Lang}}" class="{{.Class}}">
<head>
    <meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.Title}}</title>
	<link rel="stylesheet" href="{{.CSS}}" />
	{{range .MetaTags}}
		<meta name="{{.Name}}" content="{{.Content}}" property="{{.Property}}" />
	{{end}}
	{{range .Links}}
		<link rel="{{.Rel}}" href="{{.Href}}" />
	{{end}}
</head>
<body>
    <div id="page">{{.RenderedContent}}</div>
	{{if .Hydrate}}
	<script type="module" src="{{.JS}}"></script>
	<script>window.PAGE_PROPS = {{.InitialProps}};</script>
	{{end}}

	{{if .IsDev}}
	<script>
      function debounce(func, timeout = 500) {
        let timer;
        return (...args) => {
          clearTimeout(timer);
          timer = setTimeout(() => {
            func.apply(this, args);
          }, timeout);
        };
      }
      
      const reload = debounce(() => {
        console.log("reloading...");
        window.location.reload(true);
      });
      
      function start() {
        let socket = new WebSocket("ws://127.0.0.1:8080/ws");
      
        socket.onmessage = reload
      
        socket.onclose = () => {
          socket = null;
          setTimeout(start, 1000);
        };
      }
      
      start();
	</script>
	{{end}}
</body>
</html>`

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

func backendOptions(page string) esbuild.BuildOptions {
	pageDir := path.Dir(page)
	pageName := path.Base(page)
	outfile := strings.TrimSuffix(path.Join(CacheDir, page), filepath.Ext(page)) + ".ssr.js"

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
		MinifyWhitespace:  !IsDev(),
		MinifyIdentifiers: !IsDev(),
		MinifySyntax:      !IsDev(),
	}
}

func buildBackend(page string) string {
	result := esbuild.Build(backendOptions(page))

	if result.Errors != nil {
		log.Fatal("Failed to build server bundle", result.Errors)
	}

	return string(result.OutputFiles[0].Contents)
}

func BuildBackendCached(page string) string {
	cacheKey := PageCacheKey(page, "ssr.js")

	cached, err := readFile(cacheKey)
	if err == nil {
		return string(cached)
	}

	result := buildBackend(page)

	if err := os.WriteFile(cacheKey, []byte(result), 0644); err != nil {
		log.Fatal("Could not write server bundle to cache:", err)
	}

	return cacheKey
}

func clientOptions(page string) esbuild.BuildOptions {
	pageDir := path.Dir(page)
	pageName := path.Base(page)
	outfile := strings.TrimSuffix(path.Join(CacheDir, page), filepath.Ext(page)) + ".js"

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
			NewTailwindPlugin(!IsDev()),
		},
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  !IsDev(),
		MinifyIdentifiers: !IsDev(),
		MinifySyntax:      !IsDev(),
	}
}

func buildClient(page string) (string, string) {
	result := esbuild.Build(clientOptions(page))

	if result.Errors != nil {
		log.Fatal("Failed to build client bundle", result.Errors)
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

	return jsResult, cssResult
}

func BuildClientCached(page string) (string, string) {
	jsCacheKey := PageCacheKey(page, "js")
	cssCacheKey := PageCacheKey(page, "css")

	_, jsErr := readFile(jsCacheKey)
	_, cssErr := readFile(cssCacheKey)
	if jsErr == nil && cssErr == nil {
		return jsCacheKey, cssCacheKey
	}

	clientBundle, clientCSS := buildClient(page)

	if err := os.WriteFile(jsCacheKey, []byte(clientBundle), 0644); err != nil {
		log.Fatal("Could not write client bundle to cache:", err)
	}

	if err := os.WriteFile(cssCacheKey, []byte(clientCSS), 0644); err != nil {
		log.Fatal("Could not write client CSS to cache:", err)
	}

	return jsCacheKey, cssCacheKey
}

func ssrCached(page string) string {
	cacheKey := PageCacheKey(page, "html")

	cached, err := readFile(cacheKey)
	if err != nil {
		log.Fatal("Could not read html from cache:", err)
	}

	return string(cached)
}
