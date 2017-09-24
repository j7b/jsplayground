# JS Playground

This is a fork of the gopherjs playground. The main differences are the actual compiler and related functionality are as decoupled from the web interface as much as possible using ec2015 promises and the generate script is replaced with a Go program.

## Javascript

    // Compile returns a Promise that resolves to the compiled Javascript and rejects with any error(s)
    Go.Compile(source)
    // Format returns a Promise that resolves to the formatted source and rejects with errors
    Go.Format(source,imports)
    // RedirectConsole redirects standard output from GopherJS code to function(line)
    Go.RedirectConsole(function(line))
    // PackageURI sets URI for loading packages
    Go.PackageURI(uri string)

## TODO

Fix imports.json generation (methods of public types)
