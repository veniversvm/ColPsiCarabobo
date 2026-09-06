// web/src/components/psi/profile/QrCode.tsx
import { createEffect } from "solid-js";
import QRCode from "qrcode";

interface QrProps {
  url: string;
  filename?: string;
}

function QRCodeGenerator(props: QrProps) {
  let canvasRef: HTMLCanvasElement | undefined;

  createEffect(() => {
    if (canvasRef && props.url) {
      QRCode.toCanvas(
        canvasRef,
        props.url,
        {
          width: 200,
          margin: 1,
          errorCorrectionLevel: 'H', // Alta corrección para que sea más robusto al imprimir
          color: {
            dark: "#1e3a8a", 
            light: "#ffffff",
          },
        },
        (error) => {
          if (error) console.error("Error generando QR:", error);
        }
      );
    }
  });

  const downloadQR = () => {
    if (!canvasRef) return;
    
    // Convertir el canvas a imagen
    const guiUrl = canvasRef.toDataURL("image/png");
    
    // Crear un link temporal para la descarga
    const link = document.createElement("a");
    link.href = guiUrl;
    link.download = props.filename || `qr-perfil-${Date.now()}.png`;
    
    // Disparar el click y remover el elemento
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div class="flex flex-col max-w-sm items-center justify-center p-1 bg-colpsi-surface rounded-2xl border border-colpsi-border">
      <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-3">
        Código QR del Perfil
      </p>
      
      <div class="bg-white p-2 rounded-xl shadow-inner border border-colpsi-border mb-3">
        <canvas 
          ref={canvasRef} 
          class="max-w-full h-auto rounded-lg"
        ></canvas>
      </div>
      
      <button
        onClick={downloadQR}
        class="flex items-center gap-2 text-[10px] font-bold text-colpsi-blue bg-white border border-gray-200 px-3 py-1.5 rounded-full hover:bg-colpsi-blue hover:text-white transition-all shadow-sm active:scale-95"
      >
        <span>📥</span>
        Descargar QR
      </button>
    </div>
  );
}

export default QRCodeGenerator;