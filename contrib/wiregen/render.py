from pathlib import Path

from wiregen import Library, build, get_theme, load_diagram, render, validate

lib = Library.load(extra_dirs=[Path("contrib/wiregen")])
diagram = load_diagram("contrib/analog-input/pcm1808_RP2350_pcm5102A.yaml")
# issues = validate(diagram, lib)
theme = get_theme("midnight")
geo = build(diagram, lib, theme)
svg = render(geo, theme, grid=True)
open("contrib/analog-input/pcm1808_RP2350_pcm5102A.svg", "w").write(svg)
