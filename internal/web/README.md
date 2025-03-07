# Web pages

Everything is served from the public html folder (Default: `_datafiles/html/public`)

The `.html` extension can be left off of page requests. For example, `http://localhost/webclient.html` can also be accessed with `http://localhost/webclient`.

You can add your own files (images, html, css) and they will be served from here as well. For example, add a folder called `test` and put a `test.html` in there. You should be able to open it up at `http://localhost/test/test.html` or `http://localhost/test/test`

**NOTE:** `.html` files are parsed as templates, using the Go `text/template` package.

**NOTE:** files beginning with `_` such as `_header.html` cannot be directly requested. Additionally, these files are loaded into memory automatically and parsed with every page request. This is a good place to put template includes (See `_header.html`, `_footer.html` and `404.html` for an example of what this looks like.

## Template variables

There are a few template variables defined for use:

`.REQUEST` - This is an object containing the web request data. See [Request.go](https://go.dev/src/net/http/request.go) for details.

`.CONFIG` - This is an object containing the MUD config data. See [configs.go](https://github.com/Volte6/GoMud/blob/master/internal/configs/configs.go) for details.

`.STATS` - This object contains a little bit of data about the server. See [stats.go](https://github.com/Volte6/GoMud/blob/master/internal/web/stats.go#L9-L13) for details.

