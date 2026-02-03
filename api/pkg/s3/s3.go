// Package s3 proporciona la infraestructura para interactuar con servicios de
// almacenamiento de objetos compatibles con la API de AWS S3 (AWS S3, MinIO, Cloudflare R2).
package s3

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appConfig "github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// S3Client envuelve el cliente oficial de AWS SDK y el nombre del bucket
// para simplificar las operaciones de carga y recuperación de archivos.
type S3Client struct {
	Client *s3.Client
	Bucket string
}

// ConnectS3 inicializa la configuración del SDK de AWS v2 y retorna un cliente S3.
// Implementa el nuevo estándar de resolución de endpoints para evitar avisos de deprecación.
func ConnectS3() (*S3Client, error) {
	// 1. Cargamos la configuración base (Región y Credenciales)
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(appConfig.Envs.S3Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			appConfig.Envs.S3AccessKey,
			appConfig.Envs.S3SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	// 2. Instanciamos el cliente de S3 con opciones personalizadas
	// Sustituimos WithEndpointResolver por la configuración directa en s3.Options
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Si existe un endpoint personalizado (MinIO), lo configuramos
		if appConfig.Envs.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(appConfig.Envs.S3Endpoint)
		}
		// Importante: MinIO y otras soluciones locales requieren Path Style (bucket/archivo)
		// en lugar de Virtual Hosted Style (bucket.s3.amazonaws.com)
		o.UsePathStyle = true
	})

	return &S3Client{
		Client: client,
		Bucket: appConfig.Envs.S3Bucket,
	}, nil
}

// VerifyConnection realiza una verificación de disponibilidad del bucket.
// Implementa una estrategia de "Self-Healing": si el bucket no existe (común en MinIO recién iniciado),
// intenta crearlo automáticamente con permisos privados.
func (s *S3Client) VerifyConnection() {
	// Establecemos un timeout corto para la verificación de red
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Intentamos obtener los metadatos del bucket (HEAD request)
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.Bucket),
	})

	if err != nil {
		log.Printf("⚠️  S3: El bucket '%s' no fue encontrado. Intentando crear...", s.Bucket)

		// Intento de creación automática
		_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(s.Bucket),
		})

		if err != nil {
			log.Printf("❌ S3 Error: No se pudo crear ni acceder al bucket: %v", err)
			return
		}
		log.Printf("✅ S3: Bucket '%s' creado y listo para usar", s.Bucket)
	} else {
		log.Printf("✅ S3: Conexión establecida con el bucket '%s'", s.Bucket)
	}
}
