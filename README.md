Scraper - Sainsburys Code Test
==============================

Usage
-----

Test:

> go test ./...

Install:

> go get github.com/ganners/scraper

Launch:

> scraper

When prompted for a URL, try this:

> http://hiring-tests.s3-website-eu-west-1.amazonaws.com/2015_Developer_Scrape/5_products.html

If you want to use phantomjs then install phantomjs into
`/usr/local/bin/phantomjs` (or modify the path in the code). Could
configure to use flags later but it's not in use at the moment.

It also has a definition file in the commit history which will fetch from the
live Sainsburys website, it is as simple as modifying that file in any case.

About
-----

To make this more interesting, I decided that I would try to approach the
'parsing' of the HTML document in a slightly different way.

Normally the approach would be to find/build something which generates a DOM
tree, then access items you want with CSS selectors or some other API.

My thought was - could you remove the document you are parsing from the fact
that it is structured, and just care about how it exists as text? The result
would look like the opposite of a templating language, and could be created by
just copying + pasting the block of repeating text which you want to extract
from, and where you want to extract you say what you want that field to be
called.

The solution would effectively look like a templating language, but it would
operate in reverse.

I even offer filters, so the resulting value that was pulled can be seperated
by a `|` and have numerous filters appear afterwards. These can do things like
remove URL encoding, strip whitespace, uppercase and lowercase.

Sainsburys.definition
---------------------

The `sainsburys.definition` file does just this, it is was copied from the
actual source (and tidied) so anything I wanted to extract I just put into a
variable name. This is something anyone could do, you wouldn't need to be
technically minded at all.

This gets functionally lexed, and then goes through a very simple parser which
will apply the lexicons to work out what should happen at certain variables.

Main
====

This is fully pipelined, it means that data simply flows from channel to
channel in goroutines (which are workers) and so on. We spawn many workers for
particular tasks and make sure that they can operate with thread safety. The
result is a lock-free, race-free (and leak-free) program.

I believe everything specific to the Sainsburys implementation is in
the main.go and definitions files. The main source of complexity is in
the `sainsburysFormatter` function.

The `sainsburysFormatter` takes the parsed file and will generate it's
own pipeline to fetch product pages (reusing existing functions to
generate workers). It will also calculate the totals from all of the
product fields.

Given more time, I might modify the definition files to add the
complexity so that it knows certain fields should be totalled, and that
it should use a particular URL to merge in content from a child
definition.
