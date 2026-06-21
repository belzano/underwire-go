// Copyright (c) 2026 Guillaume Evrard
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// --- 1. Capturer les informations de la requête ---
		start := time.Now()

		// Lire le corps de la requête (si présent)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Erreur lors de la lecture du corps de la requête: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Réinsérer le corps dans la requête pour qu'il soit disponible pour les handlers suivants
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// --- 2. Logger les détails de la requête ---
		log.Printf(
			"[HTTP REQUEST] Method=%s | Path=%s | RemoteAddr=%s | UserAgent=%s | Body=%s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			string(body),
		)

		// --- 3. Appeler le prochain handler ---
		next.ServeHTTP(w, r)

		// --- 4. Logger la durée de la requête (optionnel) ---
		log.Printf(
			"[HTTP DURATION] Method=%s | Path=%s | Duration=%v",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
