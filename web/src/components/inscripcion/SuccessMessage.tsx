// web/src/components/inscripcion/SuccessMessage.tsx
//
// Mensaje de éxito mostrado tras enviar correctamente la solicitud de inscripción.
export function SuccessMessage() {
  return (
    <div class="max-w-2xl mx-auto bg-white p-10 rounded-3xl shadow-sm border border-green-100 text-center">
      <div class="w-20 h-20 bg-green-50 text-green-600 rounded-full flex items-center justify-center text-4xl mx-auto mb-6">
        ✓
      </div>
      <h2 class="text-2xl font-black text-colpsi-blue mb-3">
        Solicitud recibida correctamente
      </h2>
      <p class="text-gray-600 text-sm leading-relaxed">
        Hemos registrado tu solicitud de pre-inscripción en el Colegio de Psicólogos del Estado Carabobo.
        Nuestro equipo administrativo la revisará y te contactaremos en un plazo de hasta
        <strong> 5 días hábiles</strong>.
      </p>
      <p class="text-colpsi-muted text-sm mt-4">
        Si tu solicitud es aprobada, recibirás un correo con tus credenciales de acceso a la plataforma.
      </p>
    </div>
  );
}