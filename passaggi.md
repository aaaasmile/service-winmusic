Per compilare il supporto per sqlite3 occorre gcc (lancio semplicemente go run .\main.go).
=> github.com/mattn/go-sqlite3
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in %PATH%

## Compilare con gcc in windows (Working)
Sembra che il mio compiler in C:/msys64/ucrt64/bin sia troppo recente. Provo con TDM-GCC dal sito
https://jmeubank.github.io/tdm-gcc/download/

TDM-GCC non l'ho messo nel PATH per non creare dei conflitti. Quindi in powershell uso:
$env:path="C:\TDM-GCC-64\bin;" + $env:path
>gcc --version    
gcc.exe (tdm64-1) 10.3.0

## Tentativo andato a vuoto
Ho usato mysys64 (che NON funziona) in power shell con 
$env:path="C:\msys64\ucrt64\bin;" + $env:path
La versione di gcc è qui:
gcc.exe (Rev4, Built by MSYS2 project) 12.2.0

Però ottengo questo errore (stato open in github, un bug che non sembra venir risolto velocemente):
\\_\_\runtime\cgo/gcc_util.c:18: undefined reference to `__imp___iob_func'
collect2.exe: error: ld returned 1 exit status

