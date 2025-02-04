package htgo

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	esbuild "github.com/evanw/esbuild/pkg/api"
	estailwind "github.com/iamajoe/esbuild-plugin-tailwind"
	v8 "rogchap.com/v8go"
)

// [Yaffle/TextEncoderTextDecoder.js](https://gist.github.com/Yaffle/5458286)
const textEncoderPolyfill = `function TextEncoder(){} TextEncoder.prototype.encode=function(string){var octets=[],length=string.length,i=0;while(i<length){var codePoint=string.codePointAt(i),c=0,bits=0;codePoint<=0x7F?(c=0,bits=0x00):codePoint<=0x7FF?(c=6,bits=0xC0):codePoint<=0xFFFF?(c=12,bits=0xE0):codePoint<=0x1FFFFF&&(c=18,bits=0xF0),octets.push(bits|(codePoint>>c)),c-=6;while(c>=0){octets.push(0x80|((codePoint>>c)&0x3F)),c-=6}i+=codePoint>=0x10000?2:1}return octets};function TextDecoder(){} TextDecoder.prototype.decode=function(octets){var string="",i=0;while(i<octets.length){var octet=octets[i],bytesNeeded=0,codePoint=0;octet<=0x7F?(bytesNeeded=0,codePoint=octet&0xFF):octet<=0xDF?(bytesNeeded=1,codePoint=octet&0x1F):octet<=0xEF?(bytesNeeded=2,codePoint=octet&0x0F):octet<=0xF4&&(bytesNeeded=3,codePoint=octet&0x07),octets.length-i-bytesNeeded>0?function(){for(var k=0;k<bytesNeeded;){octet=octets[i+k+1],codePoint=(codePoint<<6)|(octet&0x3F),k+=1}}():codePoint=0xFFFD,bytesNeeded=octets.length-i,string+=String.fromCodePoint(codePoint),i+=bytesNeeded+1}return string};`
const processPolyfill = `var process = {env: {NODE_ENV: "production"}};`
const consolePolyfill = `var console = {log: function(){}};`

const htmlTemplate = `<!DOCTYPE html>
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
	<script type="module" src="{{.JS}}"></script>
	<script>window.PAGE_PROPS = {{.InitialProps}};</script>
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
	outfile := strings.TrimSuffix(path.Join(cacheDir, page), filepath.Ext(page)) + ".ssr.js"

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
			"js": textEncoderPolyfill + processPolyfill + consolePolyfill,
		},
		Loader: map[string]esbuild.Loader{
			".tsx": esbuild.LoaderTSX,
			".css": esbuild.LoaderEmpty,
		},
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  !isDev(),
		MinifyIdentifiers: !isDev(),
		MinifySyntax:      !isDev(),
	}
}

func buildBackend(page string) string {
	result := esbuild.Build(backendOptions(page))

	if result.Errors != nil {
		log.Fatal("Failed to build client bundle", result.Errors)
	}

	return string(result.OutputFiles[0].Contents)
}

func buildBackendCached(page string) string {
	cacheKey := pageCacheKey(page, "ssr.js")

	cached, err := readFile(cacheKey)
	if err == nil {
		return string(cached)
	}

	result := buildBackend(page)

	if err := os.WriteFile(cacheKey, []byte(result), 0644); err != nil {
		log.Fatal("Could not write client bundle to cache:", err)
	}

	return cacheKey
}

func clientOptions(page string) esbuild.BuildOptions {
	pageDir := path.Dir(page)
	pageName := path.Base(page)
	outfile := strings.TrimSuffix(path.Join(cacheDir, page), filepath.Ext(page)) + ".js"

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
			estailwind.NewTailwindPlugin(!isDev()),
		},
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  !isDev(),
		MinifyIdentifiers: !isDev(),
		MinifySyntax:      !isDev(),
	}
}

func buildClient(page string) (string, string) {
	result := esbuild.Build(clientOptions(page))

	if result.Errors != nil {
		log.Fatal("Failed to build client bundle", result.Errors)
	}

	return string(result.OutputFiles[0].Contents), string(result.OutputFiles[1].Contents)
}

func buildClientCached(page string) (string, string) {
	jsCacheKey := pageCacheKey(page, "js")
	cssCacheKey := pageCacheKey(page, "css")

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

func ssr(page string, props string) string {
	backendBundle := buildBackendCached(page)

	ctx := v8.NewContext()
	_, err := ctx.RunScript(backendBundle, "bundle.js")
	if err != nil {
		log.Fatal("Failed to evaluate bundled script:", err)
	}

	val, err := ctx.RunScript("renderPage("+props+")", "render.js")
	if err != nil {
		log.Fatal("Failed to render React component:", err)
	}

	return val.String()
}
