#!/bin/bash
# Adjust an image's opacity.

src=${1:?Must provide source image}
dst=${2:?Must provide dst image}
opacity=${3:?Must provide opacity value (percent opaque)}

convert ${src} -background white -alpha on -channel a -evaluate set ${opacity}% ${dst}
