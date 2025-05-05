# Impuls Code-Challenge

Einige der Nachbarländer (AT/BE/NL) veröffentlichen nahezu "live" den aktuellen Stand der jeweiligen Regelzone, beispielsweise in einer 1-Minuten- oder 5-Minuten-Auflösung. Diese Regelzonen der Nachbarländer korrelieren auch mit der deutschen Regelzone. Zwar ist diese Korrelation nicht direkt 1:1, jedoch bieten die Daten eine aussagekräftige Grundlage und sind für die Händler entsprechend wichtig. Daher möchten wir diese Daten in unseren Regelzonenplot in Grafana integrieren.

Für Österreich (APG) erscheinen die entsprechenden 5-Minuten-Daten besonders zügig.

[Link zur Datenquelle](<https://markttransparenz.apg.at/de/markt/Markttransparenz/Netzregelung/Deltaregelzone#:~:text=Die%20Deltaregelzone%20ist%20der%20%C3%9Cberschuss,Bilanzgruppen%2DAbweichungen%20(Ausgleichsenergie)>)

Bestehende RZ-Plots der Nachbarländer in Grafana beziehen ihre Daten derzeit von Volue, welche allerdings nicht so schnell verfügbar sind.

Entwickeln Sie eine Datenintegration unter Berücksichtigung von Best Practices, die die Daten kontinuierlich abruft, validiert und in ein einheitliches Zielformat überführt.

## Datenbank-Schema für die Speicherung von Zeitreihendaten

```sql
CREATE TABLE public.actual (
    "time" timestamp without time zone NOT NULL,
    data double precision NOT NULL,
    metric_id uuid NOT NULL REFERENCES public.metric (metric_id),
    scope_id uuid NOT NULL REFERENCES public.scope (scope_id),

    PRIMARY KEY ("time", metric_id, scope_id)
);
```
