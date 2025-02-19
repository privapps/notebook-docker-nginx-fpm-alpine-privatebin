##Nimble implementation 

It is much more lightweight than Go. Thanks Deepseek R1, ChatGPT and Claude.

Note, it took around 8 hours with AI to port the go version to nim, looks like those LLMs do not know nim very well. ChatGPT 4o-min and Gemini 2.0 Flash can not give me the basic nim http hello word server right which tooke me a lot of time and it was frustrated.

To build:
```
nim c -d:ssl --mm:arc --opt:size -d:threads --threads:on -d:NIM_MAX_THREADS=2 -d:release --passC:-flto --passL:-s httpbeast.nim
upx --best --lzma server
```

httpbeast.nim file is small ~400K, but use a lot of RSS, 25M for single thread vs. 7M for go.

