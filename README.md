[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/AntoineAugusti/avurnav)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/AntoineAugusti/avurnav/blob/master/LICENSE.md)

# AVURNAVs
This library can be used to get navigational warnings for Metropolitan France. Navigational warnings contain information about persons in distress, or objects and events that pose an immediate hazard to navigation. Navigational warnings are called [AVURNAV](https://fr.wikipedia.org/wiki/Avis_urgent_aux_navigateurs) (avis urgent aux navigateurs) in French.

It provides an HTTP client to get information directly from the Préfet Maritime websites, who publish navigational warnings for the sea region under their authority. It also offers a storage to persist to Redis navigational warnings.

## Notice
This software is available under the MIT license and was developed as part of the [Entrepreneur d'Intérêt Général program](https://entrepreneur-interet-general.etalab.gouv.fr) by the French government.
