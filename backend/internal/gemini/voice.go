package gemini

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/genai"
)

// Transcribe converts audio bytes to text using the configured chat model.
// mimeType should be the audio MIME type (e.g. "audio/webm", "audio/mp4").
func (c *Client) Transcribe(ctx context.Context, audio []byte, mimeType string) (string, error) {
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Transcribe the following audio verbatim. Respond with only the transcription text, no preamble."},
				{InlineData: &genai.Blob{MIMEType: mimeType, Data: audio}},
			},
			Role: genai.RoleUser,
		},
	}

	resp, err := c.gc.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("gemini: transcribe: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", errors.New("gemini: transcribe: no candidates in response")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}
	return strings.TrimSpace(sb.String()), nil
}

// Synthesize converts text to speech and returns WAV audio bytes.
// The returned bytes form a valid RIFF/WAV file playable by any browser <audio>.
func (c *Client) Synthesize(ctx context.Context, text string) ([]byte, error) {
	cfg := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Kore",
				},
			},
		},
	}

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{{Text: text}},
			Role:  genai.RoleUser,
		},
	}

	resp, err := c.gc.Models.GenerateContent(ctx, c.ttsModel, contents, cfg)
	if err != nil {
		return nil, fmt.Errorf("gemini: synthesize: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, errors.New("gemini: synthesize: no candidates in response")
	}

	// Find the first InlineData part (raw PCM audio).
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData == nil || len(part.InlineData.Data) == 0 {
			continue
		}
		sampleRate := parseSampleRate(part.InlineData.MIMEType, 24000)
		return pcmToWav(part.InlineData.Data, sampleRate, 1, 16), nil
	}

	return nil, errors.New("gemini: synthesize: no audio data in response")
}

// parseSampleRate extracts the sample rate from a MIME type string like
// "audio/L16;rate=24000". Returns the default if not found or unparseable.
func parseSampleRate(mimeType string, defaultRate int) int {
	for _, part := range strings.Split(mimeType, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "rate=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "rate=")); err == nil && v > 0 {
				return v
			}
		}
	}
	return defaultRate
}

// pcmToWav wraps raw PCM samples in a RIFF/WAV header so browsers can play it.
// pcm must be little-endian signed 16-bit samples.
func pcmToWav(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8
	dataSize := len(pcm)
	chunkSize := 36 + dataSize

	buf := new(bytes.Buffer)
	// RIFF chunk
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(chunkSize)) //nolint:errcheck
	buf.WriteString("WAVE")
	// fmt sub-chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))                  //nolint:errcheck // sub-chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))                   //nolint:errcheck // PCM format
	binary.Write(buf, binary.LittleEndian, uint16(channels))            //nolint:errcheck
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))          //nolint:errcheck
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))            //nolint:errcheck
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))          //nolint:errcheck
	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample))       //nolint:errcheck
	// data sub-chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize)) //nolint:errcheck
	buf.Write(pcm)
	return buf.Bytes()
}
