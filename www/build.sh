#!/bin/bash
#Fix intl-format-cache package.

    cp fix/intl-format-cache/src/* node_modules/intl-format-cache/src/
    cp fix/intl-format-cache/lib/* node_modules/intl-format-cache/lib/

#Build.
./node_modules/.bin/ember build --environment production
