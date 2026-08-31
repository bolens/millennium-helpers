# Millennium Helpers site accessibility evidence

The Pages landing page was rendered from `site/` and inspected at these viewports:

| Viewport | Evidence | Layout result |
| --- | --- | --- |
| 1440 × 1000 | [Desktop](desktop.png) | Two-column hero, full navigation, balanced diagnostic card and platform strip |
| 390 × 844 | [Mobile](mobile.png) | Single-column flow, full-width actions, compact navigation, no horizontal page overflow |
| 1440 × 6000 | [Full desktop guide](guide-desktop.png) | Platform installs, first-run checks, common tasks, help, and technical handoff |
| 390 × 6000 | [Full mobile guide](guide-mobile.png) | User guide reflows into one readable column with scrollable commands |

Lighthouse 13.4.1 reported an accessibility score of **100/100** with no failed binary audits on August 31, 2026. Separate light- and dark-mode runs covered accessible names, landmarks, heading order, link names, focusable controls, and color contrast.

The layout also provides:

- a keyboard-visible skip link and focus indicator;
- semantic header, navigation, main, section, article, list, description-list, and footer landmarks;
- hidden decorative marks and descriptive labels for diagnostic and install examples;
- reflow breakpoints at 850 and 480 CSS pixels;
- a reduced-motion mode that disables smooth scrolling and animation;
- light and dark palettes with contrast validated in the rendered page.
