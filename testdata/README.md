# testdata

## pattern.h264

10 seconds of SMPTE color bars at 25 fps, baseline H264, Annex-B bitstream. The source block loops this file over RTSP.

Generate once with ffmpeg:

```bash
ffmpeg -f lavfi -i smptebars=size=640x480:rate=25 -t 10 \
    -c:v libx264 -profile:v baseline -tune zerolatency \
    -g 25 -pix_fmt yuv420p -f h264 testdata/pattern.h264
```

The file is gitignored — regenerate locally before running tests or scenarios.
