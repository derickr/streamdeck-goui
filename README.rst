Streamdeck GO-UI
================

This repo holds the project which drives my Stream Deck application. It's not
intended for general purpose re-use, but might be useful as research for your
own integrations. It is moderately configurable though.

That said, if you want to suggest improvements, please feel free to raise
issues or pull requests, but be aware that I'm unlikely to use things unless
they're useful in the way I use StreamDeck

Uses Kevin Bowman's https://github.com/magicmonkey/go-streamdeck
and is based on Lorna Jane Mitchel's
https://github.com/lornajane/streamdeck-tricks.

Usage
-----

First copy ``config.yml.default`` to ``config.yml``. In this file, you
configure the information the Stream Deck shows.

First, define how many buttons your Stream Deck has::

	button_count: 15

If you want to change where images are stored, you can use the
``images_buttons`` setting. It's probably best to leave the default::

	images_buttons: "images/buttons"

And then you need to define pages. As I have fewer buttons, this is a nice way
of allowing more things to be done.


Each page has a name, and an image (which should be 96x96 PNG/JPG)::

	pages:
	  - Name: "OBSPage"
		Image: "obs-logo.png"
	  - Name: "Xdebug"
		Image: "xdebug.png"
	  - Name: "Home"
		Image: "house.png"
		Last: true

The ``Last`` flag makes sure to put it as the bottom-right corner. And it does
that for **every** page.

Each page is a list with buttons that you can put into a specific position,
through the ``Index`` argument.

*Some* types (such as OBS buttons) take up multiple buttons and allow you to
specify a length. The default config has this as ``Home`` page::

	Home:
	  - Index: 0
		Type: pages
	  - Index: 10
		Type: brightness
		Arguments:
		  - LevelLow: 10
		  - ImageLow: "bright-000.png"
		  - LevelNormal: 50
		  - ImageNormal: "bright-050.png"
		  - LevelHigh: 90
		  - ImageHigh: "bright-100.png"
	  - Index: 12
		Type: clock
	  - Index: 13
		Type: command
		Image: "lock.png"
		Arguments:
		  - Command: "/home/derick/bin/obs/lock.sh

A description of each type follows:

pages
	Shows from the ``Index`` all the defined pages. This can of course be
	multiple buttons.

brightness
	Adds a 'brightness' button to change the brightness of the Stream Deck.

	The button has 6 ``Arguments``, all required, and they allow you to change
	the brightness levels. I recommend you don't use 0, as that means you
	can't see what the buttons do any more!

clock
	Shows the current time.

command
	Runs a shell script defined in the ``Arguments.Command`` parameter.

obs_scenes
	If OBS is running, and the button is rendered on a new page, it loads the
	first ``Arguments.SceneCount`` scenes from OBS through its web sockets
	plugin. It starts putting scenes on the Stream Deck from the position
	defined by ``Index``, and runs ``Arguments.SceneCount`` long.

	You can configure images for each scene through the ``obs_schenes``
	configuration setting. Each schene definition is an array key name and
	``name`` that must match the OBS scene name. The ``image`` can also be
	changed.

obs_record
	Renders a record button to allow you to start recording in OBS. When
	recording is active it also highlights that this is the case.

obs_stream
	Renders a stream button to allow you to start streaming in OBS. When
	streaming is active it also highlights that this is the case.

obs_toggle_mute
	Switches the ``Arguments.Source`` between active and non active. I use
	this for my Mic. It also shows the state with the buttons defined through
	``ImageOn`` and ``ImageOff``.

home
	Switches to the "Home" page.
