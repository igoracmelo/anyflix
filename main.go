package main

import (
	"anyflix/rarbgapi"
	"anyflix/tmdbapi"
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/movie/", func(w http.ResponseWriter, r *http.Request) {
		// id := strings.TrimPrefix(r.URL.Path, "/movie/")

		// cl := NewClient()
		// m, err := cl.FindMovie(id)
		// if err != nil {
		// 	log.Print(err)
		// }
		var err error

		m := tmdbapi.Movie{
			ID:          "670292",
			Title:       "The Creator",
			ReleaseYear: 2023,
			PosterURL:   "https://www.themoviedb.org/t/p/w300_and_h450_bestv2/vBZ0qvaRxqEhZwl6LWmruJqWE8Z.jpg",
			BackdropURL: "https://www.themoviedb.org/t/p/original/kjQBrc00fB2RjHZB3PGR4w9ibpz.jpg",
			Overview:    "Amid a future war between the human race and the forces of artificial intelligence, a hardened ex-special forces agent grieving the disappearance of his wife, is recruited to hunt down and kill the Creator, the elusive architect of advanced AI who has developed a mysterious weapon with the power to end the war—and mankind itself.",
			Directors:   []string{"Gareth Edwards", "Claudio Sampaio"},
		}

		// rarbg := rarbgapi.NewClient()

		// sources, err := rarbg.Search(fmt.Sprintf("%s %d", m.Title, m.ReleaseYear), "movies", "seeders", "DESC")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// log.Printf("sources: %#+v\n", sources)

		sources := []rarbgapi.Result{rarbgapi.Result{Title: "The.Creator.2023.2160p.AMZN.WEB-DL.DDP5.1.HDR.H.265.YG⭐", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-2160p-amzn-web-dl-ddp5-1-hdr-h-265-yg-5883404.html", HSize: "14.4 GB", Resolution: 2160, Seeders: 1482, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.2160p.WEB-DL.DDP5.1.SDR.H265-AOC[TGx]", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-2160p-web-dl-ddp5-1-sdr-h265-aoc-tgx-5883595.html", HSize: "14.7 GB", Resolution: 2160, Seeders: 493, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.REPACK.2160p.MA.WEB-DL.DDP5.1.Atmos.DV.HDR.H.265-FLUX", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-repack-2160p-ma-web-dl-ddp5-1-atmos-dv-hdr-h-265-flux-5884666.html", HSize: "23.4 GB", Resolution: 2160, Seeders: 433, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.2160p.HDR10.PLUS.ENG.And.ESP.LATINO.DDP5.1.x265.MKV-BEN.THE.MEN", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-2160p-hdr10-plus-eng-and-esp-latino-ddp5-1-x265-mkv-ben-the-men-5883636.html", HSize: "16.1 GB", Resolution: 2160, Seeders: 380, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.2160p.MA.WEB-DL.DDP5.1.Atmos.DV.HDR.H.265-FLUX.", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-2160p-ma-web-dl-ddp5-1-atmos-dv-hdr-h-265-flux-5884233.html", HSize: "23.4 GB", Resolution: 2160, Seeders: 329, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator 2023 HDR 2160p WEB H265-HUZZAH", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-hdr-2160p-web-h265-huzzah-5883926.html", HSize: "14.4 GB", Resolution: 2160, Seeders: 194, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.AMZN.WEB-DL.DDP5.1.H.264.YG⭐", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-amzn-web-dl-ddp5-1-h-264-yg-5883254.html", HSize: "8.6 GB", Resolution: 1080, Seeders: 3209, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.AMZN.WEBRip.1600MB.DD5.1.x264-GalaxyRG", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-amzn-webrip-1600mb-dd5-1-x264-galaxyrg-5883613.html", HSize: "1.6 GB", Resolution: 1080, Seeders: 1933, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.AMZN.WEBRip.DDP5.1.x265.10bit-GalaxyRG265", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-amzn-webrip-ddp5-1-x265-10bit-galaxyrg265-5883798.html", HSize: "3.7 GB", Resolution: 1080, Seeders: 1686, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.HDTS.X264.Dual.YG⭐", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-hdts-x264-dual-yg-5835089.html", HSize: "4.8 GB", Resolution: 1080, Seeders: 565, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.WEB-DL.DDP5.1.H264-AOC[TGx]", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-web-dl-ddp5-1-h264-aoc-tgx-5883592.html", HSize: "8.7 GB", Resolution: 1080, Seeders: 560, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator 2023 1080p NEW HD-TS x264 AAC - HushRips", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-new-hd-ts-x264-aac-hushrips-5840263.html", HSize: "2 GB", Resolution: 1080, Seeders: 461, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator 2023 REPACK 1080p MA WEB-DL DDP5 1 Atmos H 264-FLUX", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-repack-1080p-ma-web-dl-ddp5-1-atmos-h-264-flux-5884383.html", HSize: "7.9 GB", Resolution: 1080, Seeders: 402, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.AMZN.WEB-DL.DDP5.1.H.264.Dual.YG⭐", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-amzn-web-dl-ddp5-1-h-264-dual-yg-5883843.html", HSize: "9 GB", Resolution: 1080, Seeders: 342, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-10bit-webrip-6ch-x265-hevc-psa-5885709.html", HSize: "1.4 GB", Resolution: 1080, Seeders: 231, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator (2023) (1080p MA WEB-DL x265 HEVC 10bit EAC3 Atmos 5.1 Ghost) [QxR]", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-ma-web-dl-x265-hevc-10bit-eac3-atmos-5-1-ghost-qxr-5888941.html", HSize: "6.1 GB", Resolution: 1080, Seeders: 188, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator (2023) iTA-ENG.WEBDL.1080p.x264-Dr4gon MIRCrew.mkv", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-ita-eng-webdl-1080p-x264-dr4gon-mircrew-mkv-5884434.html", HSize: "2.8 GB", Resolution: 1080, Seeders: 175, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA.", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-10bit-webrip-6ch-x265-hevc-psa-5884710.html", HSize: "1.4 GB", Resolution: 1080, Seeders: 175, Languages: []string(nil)}, rarbgapi.Result{Title: "The Creator 2023 1080p WebRip x265 KONTRAST [NikaNika]", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-1080p-webrip-x265-kontrast-nikanika-5884349.html", HSize: "2.5 GB", Resolution: 1080, Seeders: 134, Languages: []string(nil)}, rarbgapi.Result{Title: "The.Creator.2023.720p.AMZN.WEBRip.900MB.x264-GalaxyRG", URL: "https://www2.rarbggo.to//torrent/the-creator-2023-720p-amzn-webrip-900mb-x264-galaxyrg-5883610.html", HSize: "897.6 MB", Resolution: 720, Seeders: 1122, Languages: []string(nil)}}

		data := struct {
			Content tmdbapi.Movie
			Sources []rarbgapi.Result
		}{
			Content: m,
			Sources: sources,
		}

		err = template.Must(template.ParseFiles("pages/content.tmpl.html")).Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
