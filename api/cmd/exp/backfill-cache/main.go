// api/cmd/exp/backfill-cache/main.go
//
// One-off: reescribe el metadata Cache-Control de los objetos YA existentes en
// el bucket S3/MinIO. Los objetos subidos tras el cambio de pkg/s3/upload.go ya
// llevan el header; este tool lo aplica al resto sin re-subir los archivos.
//
// Uso:
//
//	go run ./cmd/exp/backfill-cache --dry-run   # previsualizar (no escribe)
//	go run ./cmd/exp/backfill-cache             # aplicar
//
// Requiere la configuración del API (env.config.go) para conectarse a S3.
package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "solo listar los objetos que se modificarían, sin escribir")
	flag.Parse()

	config.InitConfig()

	client, err := s3.ConnectS3(context.Background())
	if err != nil {
		log.Fatal().Err(err).Str("component", "backfill-cache").Msg("No se pudo conectar a S3/MinIO")
	}

	ctx := context.Background()
	var (
		scanned   int
		modified  int
		unchanged int
		errorsN   int
	)

	log.Info().
		Bool("dry_run", *dryRun).
		Str("bucket", client.Bucket).
		Str("component", "backfill-cache").
		Msg("Iniciando backfill de Cache-Control")

	var continuation *string
	for {
		listOut, err := client.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
			Bucket:            aws.String(client.Bucket),
			ContinuationToken: continuation,
		})
		if err != nil {
			log.Fatal().Err(err).Str("component", "backfill-cache").Msg("Error al listar objetos")
		}

		for _, obj := range listOut.Contents {
			key := aws.ToString(obj.Key)
			scanned++

			cc := s3.CacheControlFor(folderOf(key))

			head, err := client.Client.HeadObject(ctx, &awss3.HeadObjectInput{
				Bucket: aws.String(client.Bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				errorsN++
				log.Error().Err(err).Str("key", key).Str("component", "backfill-cache").Msg("No se pudo leer metadata del objeto")
				continue
			}

			if head.CacheControl != nil && *head.CacheControl == cc {
				unchanged++
				continue
			}

			modified++
			log.Info().
				Str("key", key).
				Str("cache_control", cc).
				Bool("dry_run", *dryRun).
				Str("component", "backfill-cache").
				Msg("Objeto a modificar")

			if *dryRun {
				continue
			}

			if _, err := client.Client.CopyObject(ctx, &awss3.CopyObjectInput{
				Bucket:             aws.String(client.Bucket),
				Key:                aws.String(key),
				CopySource:         aws.String(fmt.Sprintf("%s/%s", client.Bucket, key)),
				CacheControl:       aws.String(cc),
				ContentType:        head.ContentType,
				ContentEncoding:    head.ContentEncoding,
				ContentLanguage:    head.ContentLanguage,
				ContentDisposition: head.ContentDisposition,
				MetadataDirective:  types.MetadataDirectiveReplace,
			}); err != nil {
				errorsN++
				log.Error().Err(err).Str("key", key).Str("component", "backfill-cache").Msg("Falló el CopyObject")
			}
		}

		if listOut.IsTruncated != nil && *listOut.IsTruncated {
			continuation = listOut.NextContinuationToken
			continue
		}
		break
	}

	log.Info().
		Int("scanned", scanned).
		Int("modified", modified).
		Int("unchanged", unchanged).
		Int("errors", errorsN).
		Bool("dry_run", *dryRun).
		Str("component", "backfill-cache").
		Msg("Backfill finalizado")
}

// folderOf extrae la carpeta virtual (primer segmento) de una key S3.
func folderOf(key string) string {
	if i := strings.Index(key, "/"); i > 0 {
		return key[:i]
	}
	return ""
}
