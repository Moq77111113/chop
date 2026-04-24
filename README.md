# chop

RTSP / network impairment testbench. See `docs/superpowers/specs/` for design.

## Build

    make build

## Run smoke scenario

    ./chop run examples/smoke.yml
    # then in another terminal:
    ffplay rtsp://127.0.0.1:8501/cam1
