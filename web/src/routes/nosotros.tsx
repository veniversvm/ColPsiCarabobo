// web/src/routes/nosotros.tsx

import { A } from "@solidjs/router";
import { Title, Meta, Link } from "@solidjs/meta";

const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";

export default function AboutUs() {
  const canonicalUrl = `${SITE_URL}/nosotros`;
  const ogImage = `${SITE_URL}/og-default.jpg`;
  const pageTitle = "Nuestra Institución | Colegio de Psicólogos del Estado Carabobo";
  const pageDescription =
    "Conoce la historia, misión, visión, valores y objetivos del Colegio de Psicólogos del Estado Carabobo (CPEC). Comprometidos con la ética gremial y la salud mental en Venezuela.";

  return (
    <>
      {/* ── SEO & METADATA (SSR) ───────────────────────────────────────────── */}
      <Title>{pageTitle}</Title>
      <Meta name="description" content={pageDescription} />
      <Meta
        name="keywords"
        content="Colegio de Psicólogos, Carabobo, Venezuela, CPEC, historia, misión, visión, objetivos, valores gremiales, salud mental, directiva"
      />
      <Meta name="robots" content="index, follow" />

      <Meta property="og:type" content="website" />
      <Meta property="og:url" content={canonicalUrl} />
      <Meta property="og:title" content={pageTitle} />
      <Meta property="og:description" content={pageDescription} />
      <Meta property="og:image" content={ogImage} />
      <Meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
      <Meta property="og:locale" content="es_VE" />

      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={pageTitle} />
      <Meta name="twitter:description" content={pageDescription} />
      <Meta name="twitter:image" content={ogImage} />

      <Link rel="canonical" href={canonicalUrl} />

      {/* ── CONTENIDO DE LA PÁGINA ─────────────────────────────────────────── */}
      <main class="bg-colpsi-bg min-h-screen pb-24 font-sans">
        {/* ── HERO SUTIL ─────────────────────────────────────────────────── */}
        <header class="bg-heraldic py-20 px-6 border-b border-blue-900 shadow-inner">
          <div class="max-w-4xl mx-auto text-center">
            <div class="inline-block px-5 py-2 bg-blue-800/50 text-colpsi-yellow rounded-full text-[10px] font-black tracking-[0.2em] uppercase mb-6 border border-colpsi-yellow/30">
              Institución
            </div>
            <h1 class="text-4xl md:text-6xl font-black text-white tracking-tight leading-tight">
              Nuestra{" "}
              <span class="text-colpsi-yellow italic">Misión, Visión y Valores</span>
            </h1>
          </div>
        </header>

        <section class="max-w-4xl mx-auto px-6 py-16 space-y-16 -mt-10 relative z-10">
          {/* ── QUIÉNES SOMOS ────────────────────────────────────────────── */}
          <div class="bg-white rounded-[2.5rem] shadow-premium p-8 md:p-12 border border-colpsi-border grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
            <div class="space-y-6">
              <h2 class="text-2xl font-black text-colpsi-blue border-l-4 border-colpsi-yellow pl-4 uppercase tracking-tight">
                ¿Quiénes Somos?
              </h2>
              <p class="text-gray-600 text-base leading-relaxed font-medium text-justify">
                El Colegio de Psicólogos del Estado Carabobo (CPEC) es una
                corporación profesional con personalidad jurídica y patrimonio
                propio, que agrupa, representa y regula a los profesionales de
                la salud mental licenciados en Psicología que ejercen, residen o
                están domiciliados dentro de la jurisdicción del Estado
                Carabobo, Venezuela.
              </p>
            </div>
            <div class="bg-blue-50 p-8 rounded-3xl border border-blue-100 relative overflow-hidden">
              <div class="absolute top-0 right-0 text-9xl opacity-5 pointer-events-none -mt-4 -mr-4">
                ⚖️
              </div>
              <p class="text-colpsi-blue text-sm leading-relaxed font-bold relative z-10 text-justify">
                Como institución gremial, el CPEC no solo vigila el estricto
                cumplimiento normativo y ético de la disciplina, sino que se
                erige como un puente activo entre la ciencia psicológica y el
                bienestar biopsicosocial de la colectividad. Guiados por la
                legislación nacional y un profundo compromiso social, nos
                dedicamos a defender los derechos fundamentales de nuestros
                agremiados, promover su constante desarrollo científico y
                asegurar que la población carabobeña reciba una atención
                psicológica de la más alta calidad, sustentada en la evidencia,
                la empatía y la responsabilidad social.
              </p>
            </div>
          </div>

          {/* ── HISTORIA ─────────────────────────────────────────────────── */}
          <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-colpsi-border group hover:border-colpsi-yellow transition-colors">
            <div class="flex items-center gap-4 mb-8 border-b border-colpsi-border pb-6">
              <div class="w-14 h-14 bg-amber-50 text-amber-600 rounded-2xl flex items-center justify-center text-3xl group-hover:scale-110 transition-transform border border-amber-100">
                🏛️
              </div>
              <h3 class="text-2xl font-black text-colpsi-blue uppercase tracking-tight">
                Historia
              </h3>
            </div>

            <div class="space-y-6 text-gray-600 text-sm leading-relaxed font-medium text-justify">
              <p>
                La historia del Colegio de Psicólogos del Estado Carabobo es el
                reflejo de una lucha sostenida por la institucionalidad y el
                reconocimiento científico de la profesión en la región central
                del país. Aunque la corporación profesional ya venía funcionando
                activamente bajo una estructura organizativa previa desde el año
                1967 —liderada sucesivamente en sus albores por los psicólogos
                titulares Radamés Díaz, Tomás Eduardo Aponte Bobadilla e Iván
                Linares Alemán—, el verdadero hito de consolidación legal y
                jurídica se alcanzó a finales de la década de los setenta.
              </p>

              <div class="bg-amber-50 border-l-4 border-amber-400 p-5 rounded-r-2xl my-6">
                <p class="text-amber-900 italic font-bold">
                  El 10 de diciembre de 1978, atendiendo al mandato histórico y
                  jurídico de la recién promulgada Ley de Ejercicio de la
                  Psicología en Venezuela (del 11 de septiembre de ese mismo
                  año), un nutrido grupo de profesionales de la región se dio
                  cita en una célebre Asamblea Constituyente. El encuentro tuvo
                  lugar en la ciudad de Valencia, específicamente en la sede del
                  Colegio de Médicos del Estado Carabobo, ubicada en la
                  Urbanización Guaparo.
                </p>
              </div>

              <p>
                Esta histórica asamblea estuvo presidida de forma honorífica e
                institucional por el psicólogo Erick Becker, quien para el
                momento ejercía la presidencia de la Federación de Psicólogos de
                Venezuela (FPV). Tras certificar que se cumplían a cabalidad
                todas las exigencias legales, se declaró formalmente constituido
                el Colegio de Psicólogos del Estado Carabobo.
              </p>

              <p>
                Para dar continuidad al trabajo gremial y de conformidad con las
                disposiciones transitorias de la ley sectorial, la Junta
                Directiva que venía gestionando la asociación previa asumió de
                forma automática el mandato provisional del naciente organismo
                regulador. Aquel histórico primer Consejo Directivo de 1978
                quedó inmortalizado por los siguientes profesionales pioneros:
              </p>

              <div class="bg-colpsi-surface rounded-3xl border border-gray-200 p-6 md:p-8 my-8 shadow-inner text-left">
                <ul class="grid grid-cols-1 sm:grid-cols-2 gap-y-6 gap-x-10 text-sm">
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Presidente
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Iván Linares Alemán
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Primer Vicepresidente
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Miriam Ojeda de Lugo
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Segundo Vicepresidente
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Mauricio Sztaflijmorc
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Secretario General
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Guillermo García Natera
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Secretaria de Actas
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Holanda Zerpa de Pedrotti
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Primer Vocal
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Dalia Matute de Carugno
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Segundo Vocal
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Ocarina García de Gonzalo
                    </span>
                  </li>
                  <li class="flex flex-col border-b border-gray-200 pb-3 gap-1">
                    <span class="text-[10px] text-colpsi-muted uppercase tracking-widest font-black">
                      Tercer Vocal
                    </span>
                    <span class="font-black text-colpsi-blue text-base">
                      Ana María López de Grahan
                    </span>
                  </li>
                </ul>
              </div>

              <p>
                Desde aquel diciembre de 1978, el CPEC ha mantenido su sede en
                la ciudad de Valencia, con la facultad legal de extender
                delegaciones y dependencias a los municipios y localidades más
                cercanas de su jurisdicción para asegurar un acompañamiento
                cercano y eficiente a todos los psicólogos de la región.
              </p>
            </div>
          </div>

          {/* ── MISIÓN Y VISIÓN ──────────────────────────────────────────── */}
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div class="bg-white p-10 rounded-[2.5rem] shadow-sm border-2 border-transparent hover:border-colpsi-blue transition-all group">
              <div class="w-14 h-14 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center text-3xl mb-6 group-hover:scale-110 group-hover:rotate-3 transition-transform border border-blue-100">
                🎯
              </div>
              <h3 class="text-xl font-black text-colpsi-blue mb-4 uppercase tracking-widest">
                Misión
              </h3>
              <p class="text-gray-600 leading-relaxed text-sm font-medium space-y-4 text-justify">
                <span class="block">
                  Garantizar la ordenación, defensa y salvaguarda del ejercicio
                  legal y ético de la psicología en el Estado Carabobo, velando
                  por la dignificación socioeconómica y científica de sus
                  agremiados.
                </span>
                <span class="block">
                  Nos comprometemos a promover el avance continuo de la ciencia
                  psicológica, la salud mental y el bienestar integral de las
                  comunidades, asegurando una praxis profesional humanista,
                  accesible y estrictamente apegada a los principios éticos, de
                  cara a las realidades y necesidades de nuestra sociedad.
                </span>
              </p>
            </div>

            <div class="bg-white p-10 rounded-[2.5rem] shadow-sm border-2 border-transparent hover:border-colpsi-yellow transition-all group">
              <div class="w-14 h-14 bg-yellow-50 text-colpsi-yellow rounded-2xl flex items-center justify-center text-3xl mb-6 group-hover:scale-110 group-hover:-rotate-3 transition-transform border border-yellow-100">
                🔭
              </div>
              <h3 class="text-xl font-black text-colpsi-blue mb-4 uppercase tracking-widest">
                Visión
              </h3>
              <p class="text-gray-600 leading-relaxed text-sm font-medium space-y-4 text-justify">
                <span class="block">
                  Consolidarse como un colegio profesional de vanguardia,
                  referente a nivel nacional por su excelencia institucional, su
                  rigurosidad ética y su activo liderazgo social.
                </span>
                <span class="block">
                  Aspiramos a ser un espacio inclusivo, democrático y solidario
                  que potencie el desarrollo científico y profesional del
                  psicólogo carabobeño, posicionando la salud mental como un
                  derecho humano fundamental e indispensable para el progreso y
                  la armonía de la sociedad venezolana.
                </span>
              </p>
            </div>
          </div>

          {/* ── OBJETIVOS INSTITUCIONALES ────────────────────────────────── */}
          <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-colpsi-border group hover:border-colpsi-yellow transition-colors">
            <div class="flex items-center gap-4 mb-8 border-b border-colpsi-border pb-6">
              <div class="w-14 h-14 bg-emerald-50 text-emerald-600 rounded-2xl flex items-center justify-center text-3xl group-hover:scale-110 transition-transform border border-emerald-100">
                🚀
              </div>
              <h3 class="text-2xl font-black text-colpsi-blue uppercase tracking-tight">
                Objetivos Institucionales
              </h3>
            </div>

            {/* Resumen Visual */}
            <ul class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm font-medium text-gray-600 mb-8">
              <li class="flex items-start gap-3 bg-colpsi-surface p-5 rounded-2xl border border-colpsi-border/50">
                <span class="text-emerald-500 font-black text-lg mt-0.5">✓</span>
                Regular y vigilar el ejercicio ético y legal de la psicología en
                Carabobo.
              </li>
              <li class="flex items-start gap-3 bg-colpsi-surface p-5 rounded-2xl border border-colpsi-border/50">
                <span class="text-emerald-500 font-black text-lg mt-0.5">✓</span>
                Fomentar la actualización científica y el desarrollo profesional
                continuo.
              </li>
              <li class="flex items-start gap-3 bg-colpsi-surface p-5 rounded-2xl border border-colpsi-border/50">
                <span class="text-emerald-500 font-black text-lg mt-0.5">✓</span>
                Defender los derechos laborales, sociales y gremiales de los
                agremiados.
              </li>
              <li class="flex items-start gap-3 bg-colpsi-surface p-5 rounded-2xl border border-colpsi-border/50">
                <span class="text-emerald-500 font-black text-lg mt-0.5">✓</span>
                Promover la salud mental y el bienestar biopsicosocial en la
                comunidad.
              </li>
            </ul>

            {/* Texto Legal Extendido */}
            <div class="space-y-5 text-gray-600 text-sm leading-relaxed font-medium pt-6 border-t border-colpsi-border text-justify">
              <p>
                Derivados directamente de nuestra Acta Constitutiva, la Ley de
                Ejercicio de la Psicología y el Código de Ética Profesional, los
                fines fundamentales del Colegio son:
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Agrupación y Representación:
                </strong>{" "}
                Reunir e incorporar activamente a todos los profesionales de la
                psicología que residan o ejerzan en el Estado Carabobo,
                fortaleciendo el tejido gremial, la fraternidad y la mutua
                cooperación entre los colegas.
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Regulación del Ejercicio:
                </strong>{" "}
                Velar por el cumplimiento irrestricto de la Ley de Ejercicio de
                la Psicología, el Código de Ética del Psicólogo, los reglamentos
                internos y las directrices emanadas de la Federación de
                Psicólogos de Venezuela (FPV).
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Vigilancia Deontológica:
                </strong>{" "}
                Ejercer, a través del Tribunal Disciplinario y sus órganos
                correspondientes, la potestad de supervisar, investigar y
                sancionar las conductas que infrinjan la ética y el decoro
                profesional, protegiendo así a la ciudadanía de la mala praxis o
                el intrusismo.
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Dignificación y Previsión Social:
                </strong>{" "}
                Impulsar mecanismos robustos de bienestar socioeconómico,
                seguridad social, recreación y apoyo mutuo para los agremiados y
                sus familias, velando por condiciones laborales dignas y justas.
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Desarrollo Científico y Académico:
                </strong>{" "}
                Promover la actualización continua, la investigación científica
                y la formación académica especializada de los psicólogos en
                Carabobo, organizando congresos, talleres y espacios de
                discusión de alto nivel.
              </p>
              <p>
                <strong class="text-colpsi-blue">
                  Compromiso Comunitario:
                </strong>{" "}
                Servir como órgano consultivo en materias de salud mental
                pública, diseñando, ejecutando y apoyando programas de
                intervención comunitaria que den respuesta a los problemas
                psicosociales prioritarios de la población carabobeña.
              </p>
            </div>
          </div>

          {/* ── VALORES ────────────────────────────────────────────────────── */}
          <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-colpsi-border group hover:border-colpsi-yellow transition-colors">
            <div class="flex items-center gap-4 mb-8 border-b border-colpsi-border pb-6">
              <div class="w-14 h-14 bg-purple-50 text-purple-600 rounded-2xl flex items-center justify-center text-3xl group-hover:scale-110 transition-transform border border-purple-100">
                💎
              </div>
              <h3 class="text-2xl font-black text-colpsi-blue uppercase tracking-tight">
                Valores
              </h3>
            </div>

            <div class="space-y-8 text-gray-600 text-sm leading-relaxed font-medium">
              <p class="text-justify">
                La cultura institucional del Colegio de Psicólogos del Estado
                Carabobo descansa sobre pilares fundamentales distribuidos en
                tres dimensiones esenciales: valores éticos-científicos, valores
                humanitarios y valores gremiales.
              </p>

              {/* Bloque A */}
              <div>
                <h4 class="text-sm font-black text-colpsi-blue uppercase tracking-widest mb-4 flex items-center gap-2">
                  <span class="w-2 h-2 bg-colpsi-yellow rounded-full"></span> A.
                  Valores Éticos y Científicos
                </h4>
                <ul class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs text-justify">
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Integridad y Excelencia:
                    </strong>{" "}
                    Promovemos una práctica profesional intachable, basada en la
                    honestidad, la transparencia y el uso estricto de
                    metodologías científicas avaladas por la disciplina.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Responsabilidad Deontológica:
                    </strong>{" "}
                    Respeto y acatamiento absoluto a los deberes éticos, el
                    secreto profesional y la confidencialidad, entendiendo que
                    la confianza es la base de la relación terapéutica e
                    investigativa.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Rigurosidad y Veracidad:
                    </strong>{" "}
                    Fomento permanente de la actualización en el conocimiento
                    psicológico, evitando la difusión de teorías falsas,
                    pseudociencias o prácticas sin sustento empírico.
                  </li>
                </ul>
              </div>

              {/* Bloque B */}
              <div>
                <h4 class="text-sm font-black text-colpsi-blue uppercase tracking-widest mb-4 flex items-center gap-2 mt-8">
                  <span class="w-2 h-2 bg-colpsi-yellow rounded-full"></span> B.
                  Valores Humanitarios (Orientados al Bien Común y el Servicio)
                </h4>
                <ul class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs text-justify">
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Empatía y Compasión:
                    </strong>{" "}
                    Sensibilidad profunda ante el sufrimiento humano y las
                    dificultades psicosociales, orientando nuestras capacidades
                    a aliviar el dolor mental y promover la resiliencia en los
                    individuos.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Respeto a la Dignidad Humana:
                    </strong>{" "}
                    Reconocimiento del valor intrínseco, la autonomía y los
                    derechos humanos de cada persona, sin distinción de raza,
                    credo, orientación, condición socioeconómica o filiación
                    política.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Inclusión y Justicia Social:
                    </strong>{" "}
                    Compromiso activo con la equidad, llevando la atención y la
                    promoción de la salud mental a las poblaciones más
                    vulnerables, históricamente desatendidas o en situaciones de
                    riesgo humanitario.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">Altruismo:</strong>{" "}
                    Vocación inquebrantable de servicio social, donde la ciencia
                    y la profesión se colocan desinteresadamente a la orden del
                    desarrollo comunitario y la reconstrucción del tejido
                    social.
                  </li>
                </ul>
              </div>

              {/* Bloque C */}
              <div>
                <h4 class="text-sm font-black text-colpsi-blue uppercase tracking-widest mb-4 flex items-center gap-2 mt-8">
                  <span class="w-2 h-2 bg-colpsi-yellow rounded-full"></span> C.
                  Values Gremiales (Orientados a la Vida Institucional y al
                  Agremiado)
                </h4>
                <ul class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs text-justify">
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Solidaridad Gremial:
                    </strong>{" "}
                    Fomento del apoyo mutuo, el compañerismo y la fraternidad
                    entre los psicólogos del estado Carabobo, entendiendo que el
                    crecimiento del colectivo fortalece el crecimiento
                    individual.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Institucionalidad y Democracia:
                    </strong>{" "}
                    Respeto absoluto a la estructura organizativa, las
                    decisiones de la Asamblea General y los canales regulares
                    establecidos en nuestra Acta Constitutiva y reglamentos.
                    Creemos en la alternabilidad, el debate de ideas y la
                    participación activa.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Defensa y Dignificación Laboral:
                    </strong>{" "}
                    Compromiso constante con la defensa de los derechos
                    profesionales y las justas reivindicaciones económicas de
                    los psicólogos, combatiendo firmemente el ejercicio ilegal
                    de la profesión y la explotación laboral.
                  </li>
                  <li class="bg-colpsi-surface p-5 rounded-xl border border-colpsi-border">
                    <strong class="text-colpsi-text block mb-1">
                      Sentido de Pertenencia Institucional:
                    </strong>{" "}
                    Orgullo por nuestra historia, respeto a los fundadores y
                    mentores que trazaron el camino de la psicología carabobeña,
                    e impulso a las nuevas generaciones para que cuiden el
                    prestigio de nuestra corporación profesional.
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </section>

        {/* ── FOOTER SUTIL ─────────────────────────────────────────────────── */}
        <div class="max-w-4xl mx-auto px-6 text-center mt-6 flex flex-col sm:flex-row items-center justify-center gap-4">
          <A
            href="/directorio"
            class="inline-flex items-center justify-center gap-3 bg-colpsi-yellow text-colpsi-blue font-black px-8 py-4 rounded-2xl hover:scale-105 active:scale-95 transition-all uppercase text-sm tracking-widest shadow-lg shadow-yellow-500/20"
          >
            Consultar Directorio de Profesionales <span>→</span>
          </A>
          <A
            href="/documentos"
            class="inline-flex items-center justify-center gap-3 bg-blue-50 text-colpsi-blue font-black px-8 py-4 rounded-2xl hover:bg-blue-100 active:scale-95 transition-all uppercase text-sm tracking-widest border border-blue-100"
          >
            Marco Legal y Normativo <span>→</span>
          </A>
        </div>
      </main>
    </>
  );
}