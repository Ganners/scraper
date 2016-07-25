Scraper
=======

Usage
-----

Install:

> go get github.com/ganners/scraper

Launch:

> Scraper

When prompted for a URL, try some of these:

> http://www.sainsburys.co.uk/shop/gb/groceries/fruit-veg/ripe---ready
> http://www.sainsburys.co.uk/shop/gb/groceries/fruit-veg/melon-pineapple-kiwi

If you want to use phantomjs then install phantomjs into
`/usr/local/bin/phantomjs` (or modify the path in the code). Could
configure to use flags later but it's not in use at the moment.

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

The `sainsburys.definition` or `sainsburys-cache.definition` file does just
this, it is was copied from the actual source (and tidied) so anything I wanted
to extract I just put into a variable name. This is something anyone could do,
you wouldn't need to be technically minded at all.

This gets functionally lexed, and then goes through a very simple parser which
will apply the lexicons to work out what should happen at certain variables.

Main
====

This is fully pipelined, it means that data simply flows from channel to
channel in goroutines (which are workers) and so on. We spawn many workers for
particular tasks and make sure that they can operate with thread safety. The
result is a lock-free, race-free (and leak-free) program.

The scraping of data which is populated via JavaScript lead me down a few
routes, and I've left those implementers of the `WebReader` in. One example
attempt was to use PhantomJS but it didn't give the desired response though I'm
sure it could with a new or different package.

In the end I settled on grabbing the cached index from Google and using that
instead.  Probably not a good long-term solution and it would fail if we wanted
to scrape in realtime (we'd have to wait for Google's index to update).
