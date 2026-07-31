// web/src/lib/geo.ts
// Catálogos geográficos de Venezuela usados por los formularios del panel.
// Reflejan las listas de validación del backend (api/internal/utils/geo_venezuela.go).

export const MUNICIPIOS_CARABOBO: string[] = [
  "Bejuma",
  "Carlos Arvelo",
  "Diego Ibarra",
  "Guacara",
  "Juan José Mora",
  "Libertador",
  "Los Guayos",
  "Miranda",
  "Montalbán",
  "Naguanagua",
  "Puerto Cabello",
  "San Diego",
  "San Joaquín",
  "Valencia",
];

// Excluye Carabobo: la ubicación base se gestiona por separado (Presencia en Carabobo).
export const ESTADOS_VENEZUELA: string[] = [
  "Amazonas",
  "Anzoátegui",
  "Apure",
  "Aragua",
  "Barinas",
  "Bolívar",
  "Cojedes",
  "Delta Amacuro",
  "Dependencias Federales",
  "Distrito Capital",
  "Falcón",
  "Guárico",
  "Lara",
  "La Guaira",
  "Mérida",
  "Miranda",
  "Monagas",
  "Nueva Esparta",
  "Portuguesa",
  "Sucre",
  "Táchira",
  "Trujillo",
  "Yaracuy",
  "Zulia",
];
