# Auditoría defensiva de incidentes de software abierto — octubre de 2026

**Fecha de corte:** 25 de octubre de 2026  
**Alcance:** revisión documental de avisos públicos, trackers de distribuciones, commits, issues y PRs de mantenedores.  
**Estado:** primera pasada; no es un aviso oficial de seguridad ni una auditoría de código completa.

> **Nota de metodología:** se utilizaron herramientas automatizadas para localizar y resumir fuentes públicas; las referencias se enlazan para facilitar su verificación. No se ejecutaron muestras, no se atacaron sistemas, no se reprodujeron vulnerabilidades y no se modificó código upstream. No se ha confirmado una vulnerabilidad nueva independiente a partir de esta revisión. Este documento es una nota de investigación personal y no representa a Ollama ni a los proyectos citados.

## Resumen ejecutivo

Los incidentes divulgados alrededor del 21 de octubre de 2026 no constituyen un único ataque al kernel Linux. La evidencia pública documenta varios asuntos independientes: compromiso de la cadena de publicación de LiteLLM, cuatro vulnerabilidades de BIND 9, dos vulnerabilidades críticas de PHP y una tercera divulgada después, una vulnerabilidad de GitLab, un fallo en musl y un incidente de claves intermedias TLS de eMudhra. CVE-2026-74580 del kernel vhost se incluye únicamente como contexto: fue publicado el 21 de agosto de 2026, no el 21 de octubre.

## Hallazgos y acciones recomendadas

### P1 — LiteLLM 1.107.0: compromiso de cadena de suministro

- **Evidencia:** [informe canónico de la cadena de ataque, publicado el 23 de octubre](https://docs.litellm.ai/blog/security-incident-oct-21-2026-attack-chain) y [issue público #47826](https://github.com/BerriAI/litellm/issues/47826).
- **Hechos publicados:** el artefacto `litellm==1.107.0` y algunas etiquetas de imágenes Docker incluyeron contenido malicioso el 21 de octubre y fueron retirados/reasignados. La fuente canónica explica que una clave maestra quedó expuesta en un archivo de configuración del repositorio público `LiteLLM-Labs/lite-harness` en mayo de 2026 y no se revocó; el acceso posterior permitió extraer tokens de GitHub y alterar el proceso de publicación. La etiqueta oficial `v1.107.0` no contiene los archivos maliciosos del artefacto publicado.
- **Acción:** los usuarios que ejecutaron los artefactos afectados deben seguir el procedimiento oficial, reconstruir entornos e imágenes desde una fuente verificada y rotar los secretos a los que el entorno podía acceder. Los mantenedores deben auditar la cadena de build/publicación, las credenciales expuestas, las protecciones de ramas, la procedencia y la aprobación de releases.
- **No hacer:** no reinstalar en el entorno potencialmente afectado antes de seguir las instrucciones oficiales; no abrir un issue duplicado que repita el compromiso ya documentado. El repositorio privado de publicación no fue accesible en esta revisión.

### P1 — BIND 9: cuatro vulnerabilidades divulgadas el 21 de octubre

- **Avisos:** [CVE-2026-19630](https://kb.isc.org/docs/cve-2026-19630), [CVE-2026-94300](https://kb.isc.org/docs/cve-2026-94300), [CVE-2026-96888](https://kb.isc.org/docs/cve-2026-96888), [CVE-2026-97313](https://kb.isc.org/docs/cve-2026-97313); [aviso de lanzamiento ISC](https://lists.isc.org/pipermail/bind-announce/2026-October/001306.html).
- **Estado upstream:** ISC publicó BIND 9.20.30 y 9.21.27 con correcciones. El aviso de CVE-2026-94300 describe corrupción de memoria bajo condiciones específicas y declara que ISC no conocía exploits activos en el momento del aviso.
- **Acción:** actualizar mediante el canal de seguridad de la distribución y comprobar el paquete instalado; los números de versión y el estado de backport varían entre Debian, Ubuntu y otros proveedores. No crear un duplicado upstream sin nueva evidencia de una rama o paquete que siga vulnerable.

### P1 — PHP: CVE-2026-109618 y CVE-2026-109619

- **Avisos:** [PHP, 21 de octubre](https://www.php.net/archive/2026.php#2026-10-21-5) y [CERT-EU 2026-018](https://www.cert.europa.eu/publications/security-advisories/2026-018).
- **Versiones upstream corregidas para estos dos CVE:** PHP 8.5.12, 8.4.27, 8.3.36 y 8.2.35. Los paquetes de distribución pueden incluir backports con numeración distinta.
- **Actualización importante:** el 23 de octubre CERT-EU añadió CVE-2026-115560, otra vulnerabilidad RCE para la que aún no había detalles técnicos completos en la fecha de corte. Como medida temporal hasta una actualización ordinaria, CERT-EU recomienda desactivar `session.upload_progress.enabled` y reiniciar todos los workers.
- **Regresiones ya registradas:** [GH-24492](https://github.com/php/php-src/issues/24492) describe terminaciones de proceso con código 137 en algunas rutas de asignación de memoria después de la mitigación de seguridad. Existen PRs abiertos que cubren distintos subconjuntos: [#24548](https://github.com/php/php-src/pull/24548), [#24596](https://github.com/php/php-src/pull/24596), [#24618](https://github.com/php/php-src/pull/24618) y [#24516](https://github.com/php/php-src/pull/24516). [GH-24493](https://github.com/php/php-src/issues/24493) documenta otra posible regresión relacionada con la pérdida de sesiones.
- **Acción:** actualizar a la versión corregida o al paquete del proveedor, seguir los avisos de PHP para CVE-2026-115560 y dar seguimiento a los issues/PRs existentes. No abrir duplicados ni presentar GH-24492 automáticamente como otro RCE.

### P1 — GitLab: CVE-2026-110264

- **Aviso:** [GitLab Critical Patch Release](https://docs.gitlab.com/releases/patches/patch-release-gitlab-19-5-1-released/) y [CERT-EU 2026-019](https://www.cert.europa.eu/publications/security-advisories/2026-019/).
- **Descripción:** desbordamiento entero en el compilador de expresiones regulares que, bajo ciertas condiciones, podía permitir ejecución de código no autenticada. GitLab lo calificó CVSS 10.0.
- **Versiones corregidas publicadas:** 18.11.13, 19.0.10, 19.1.9, 19.2.9, 19.3.6, 19.4.3 y 19.5.1.
- **Acción:** actualizar la instancia autogestionada a la versión corregida del canal instalado y seguir las detecciones de explotación que recomienda GitLab. El repositorio canónico reside en GitLab.com; en esta revisión solo se contrastaron avisos públicos, no se inspeccionó código de ese repositorio.

### P1 — eMudhra: claves de certificados intermedios TLS

- **Fuentes:** [Mozilla Security Blog](https://blog.mozilla.org/security/2026/10/20/protecting-firefox-users-following-the-compromise-of-emudhra-intermediate-certificates/), [declaración de eMudhra](https://emudhra.com/en/post/certificate-security-incident), [Red Hat RHSB-2026-012](https://access.redhat.com/security/vulnerabilities/RHSB-2026-012).
- **Hechos y límites:** se reportaron ocho claves de certificados intermedios criptográficamente débiles/comprometidas. Mozilla bloqueó los certificados afectados en Firefox mediante OneCRL y anunció una actualización de NSS. eMudhra indicó que no se emitieron certificados de entidad final con los intermedios afectados y que no encontró evidencias de compromiso de su infraestructura. No debe inferirse que se confirmaron certificados fraudulentos o que todos los usuarios resultaron afectados.
- **Correcciones observadas:** CPython fusionó un bloqueo de claves SPKI en [PR #160042](https://github.com/python/cpython/pull/160042); Node.js fusionó una protección equivalente en [PR #66871](https://github.com/nodejs/node/pull/66871).
- **Acción:** comprobar por separado el almacén de confianza del sistema operativo y los almacenes embebidos en runtimes, aplicaciones e imágenes de contenedor; aplicar los updates oficiales correspondientes.

### P2 — musl: CVE-2026-107891

- **Fuentes:** [divulgación OSS-Security](https://openwall.com/lists/oss-security/2026/10/21/18), [release de Alpine](https://www.alpinelinux.org/posts/Alpine-3.21.9-3.22.7-3.23.7-3.24.3-released.html), [commit Alpine de corrección](https://github.com/alpinelinux/aports/commit/e215be1a82600bedcd72c721061ea2bf20a0b1c3).
- **Descripción:** desbordamientos enteros en `regcomp` que pueden causar asignaciones insuficientes y escrituras fuera de límites si una aplicación procesa expresiones regulares no confiables.
- **Estado observado:** Alpine publicó 3.21.9, 3.22.7, 3.23.7 y 3.24.3 con correcciones. Esto no prueba que todas las distribuciones ya hayan actualizado sus paquetes.
- **Acción:** aplicar el update del proveedor y comprobar las versiones del paquete; no abrir un duplicado cuando el parche ya está registrado.

### Contexto adicional — Linux kernel vhost: CVE-2026-74580

- **Fuentes:** [Red Hat](https://access.redhat.com/security/cve/cve-2026-74580), [Ubuntu](https://ubuntu.com/security/CVE-2026-74580), [Debian](https://security-tracker.debian.org/tracker/CVE-2026-74580).
- Este fallo afecta al caché de metadatos vring/IOTLB en vhost. Fue publicado el **21 de agosto de 2026**, no el 21 de octubre; se incluye como contexto de seguridad del kernel, no como evidencia de un ataque ocurrido el 21 de octubre.
- **Acción:** verificar la versión del paquete y los backports de la distribución. No publicar un reporte duplicado del CVE ya corregido upstream.

## Límites, prioridades y siguientes pasos

1. Actualizar sistemas y runtimes vulnerables usando el proveedor de cada distribución, comprobando el paquete exacto y su estado de backport.
2. En LiteLLM, tratar la exposición como un incidente de credenciales y cadena de suministro, no como un defecto demostrado en la etiqueta oficial del código fuente.
3. En PHP, seguir las correcciones y pruebas ya abiertas de GH-24492/GH-24493 y revisar los avisos de CVE-2026-115560.
4. En PKI, verificar almacenes de confianza embebidos además de los del sistema.
5. Solo presentar nuevos issues/PRs si una revisión de código o reproducción aislada demuestra un fallo independiente que no esté cubierto por avisos, issues o PRs existentes.

**Trabajo no realizado:** no se compilaron los proyectos, no se ejecutaron tests, no se probó un sistema afectado concreto y no se reprodujeron los fallos. Los estados de paquetes y PRs son una instantánea documental a 25 de octubre de 2026 y pueden cambiar. Esta revisión no identifica una vulnerabilidad nueva confirmada que justifique una nueva divulgación upstream.

## Fuentes primarias principales

- [LiteLLM: cadena de ataque del release v1.107.0](https://docs.litellm.ai/blog/security-incident-oct-21-2026-attack-chain)
- [ISC: lanzamientos BIND 9.20.30 y 9.21.27](https://lists.isc.org/pipermail/bind-announce/2026-October/001306.html)
- [PHP: aviso de seguridad del 21 de octubre](https://www.php.net/archive/2026.php#2026-10-21-5)
- [CERT-EU: vulnerabilidades PHP y CVE-2026-115560](https://www.cert.europa.eu/publications/security-advisories/2026-018)
- [GitLab: parche crítico 19.5.1 y ramas relacionadas](https://docs.gitlab.com/releases/patches/patch-release-gitlab-19-5-1-released/)
- [Mozilla: respuesta al incidente de eMudhra](https://blog.mozilla.org/security/2026/10/20/protecting-firefox-users-following-the-compromise-of-emudhra-intermediate-certificates/)
- [Alpine: lanzamientos del 21 de octubre](https://www.alpinelinux.org/posts/Alpine-3.21.9-3.22.7-3.23.7-3.24.3-released.html)
