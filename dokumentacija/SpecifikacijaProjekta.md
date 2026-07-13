# Specifikacija projekta Softverski agenti 2026

## Tim

Nikola Šehovac, broj indeksa: **RA115/2020**.

Repozitorijum: https://github.com/seli0001/softverski-agenti-2026

## Zadatak

Projekat radim za ocenu 10. To obuhvata: generički aktorski radni okvir
napisan od nule u Go-u, dodatne funkcionalnosti okvira, klasterovanje kroz
peer-to-peer i provider režim, upotrebu CRDT struktura i aplikaciju za
federativno učenje izgrađenu nad tim okvirom.

## Aktorski radni okvir

Okvir je generički i nije vezan za federativno učenje. FL aplikacija ga
koristi kao i svaki drugi korisnik.

Obavezni elementi:

- aktori sa asinhronim slanjem i primanjem poruka
- sanduče: baferovan kanal, po jedna gorutina po aktoru; stanje aktora dira
  samo njegova gorutina, pa nema zaključavanja
- menjanje ponašanja aktora (Become)
- lifecycle događaji kao obične poruke: Started, Stopping, Terminated, uz
  mogućnost praćenja tuđeg gašenja (Watch)
- udaljena komunikacija: TCP transport, gob serijalizacija, lokacijska
  transparentnost (PID nosi adresu čvora, pa Send sam bira lokalnu ili
  mrežnu dostavu)

Dodatne urađene funkcionalnosti:

- supervizija: hijerarhija roditelj i dete, hvatanje panike, odluke
  Resume / Stop / Restart (restart preko producer funkcije)
- persistencija stanja: event sourcing sa žurnalom (u memoriji i u fajlu)
  i snapshot mehanizmom
- SSL: uzajamni TLS na transportu, sa sopstvenim CA; čvor bez ispravnog
  sertifikata biva odbijen već pri rukovanju
- klasterovanje (opisano ispod)

## Klasterovanje

Oba režima su implementirana, a režim se bira pri pokretanju čvora.

**Peer-to-peer:** HyParView protokol. Novi čvor ulazi preko kontakt čvora
porukom Join, a vest o njemu šeta mrežom nasumičnim skokovima sa TTL-om
(ForwardJoin), pa veze novog čvora završe raspršene po klasteru. Svaki čvor
drži mali aktivni pogled (3 suseda sa kojima stvarno komunicira) i pasivni
pogled (10 rezervi). Ulazak u tuđi aktivni pogled ide kroz rukovanje
porukama NeighborRequest i NeighborResponse, pri čemu zahtev sme da bude
odbijen, osim zahteva visokog prioriteta; raskid veze ide kroz Disconnect.
Otkazi se otkrivaju periodičnim ping/pong porukama. Timeout je realizovan
kao poruka koju aktor šalje sam sebi sa zakašnjenjem (SendLater), pa mrtav
sused biva zamenjen rezervom iz pasivnog pogleda.

**Provider:** centralni registar realizovan kao aktor. Članovi se
registruju porukom Register, šalju periodične otkucaje (Heartbeat), a
registar objavljuje ažuran spisak članova (Members) i izbacuje one koji se
ne javljaju. Ovaj režim je jednostavniji i svi članovi imaju identičan
spisak, ali je registar jedina tačka otkaza, za razliku od peer-to-peer
režima koji nema centralnu komponentu.

Ceo membership sloj je napisan kao obični aktori okvira; protokoli ne
zahtevaju nikakvu posebnu infrastrukturu.

## CRDT

Implementirani tipovi: G-Counter (poseban slot za svaki čvor, merge uzima
max po slotu), PN-Counter (dva G-Countera, vrednost je razlika) i G-Set
(merge je unija). Sinhronizacija ide anti-entropy gossip razmenom preko
HyParView suseda: čvor periodično šalje celo svoje stanje, primalac radi
merge. Pošto je merge komutativan, asocijativan i idempotentan, protokol
podnosi gubitak, dupliranje i mešanje redosleda poruka. U FL aplikaciji
G-Counter broji ukupan broj obavljenih lokalnih treninga u celom klasteru.

## Federativno učenje

**Problem:** predikcija da li će let kasniti. Motivacija za federaciju:
svaka aviokompanija ima svoje operativne podatke i ne deli ih sa
konkurencijom, ali svima odgovara zajednički model. Podaci ostaju kod
vlasnika, razmenjuju se samo težine modela.

**Skup podataka:** OpenML "Airlines", 539.383 stvarna leta. Ulazni
atributi: vreme polaska i trajanje leta (normalizovani) i dan u nedelji
(one-hot kodiran), ukupno 9 ulaza. Ciljna promenljiva: kasni / ne kasni
(45% / 55%).

**Model i algoritam:** logistička regresija trenirana stohastičkim
gradijentnim spustom, sopstvena implementacija. Federacija je urađena u dva
oblika, po jedan za svaki klaster režim:

- provider režim koristi FedAvg (McMahan, 2017): agregator drži globalne
  težine i šalje ih trenerima; svaki trener ih dotrenira na svom komadu
  podataka i vrati zajedno sa brojem primera; agregator računa ponderisani
  prosek Σ(wᵢ·Nᵢ)/ΣNᵢ i pokreće sledeću rundu
- peer-to-peer režim koristi gossip learning: nema rundi ni koordinatora;
  čvor stalno trenira svoj model na svojim podacima, periodično ga šalje
  HyParView susedima, a primalac uprosečava svoj i primljeni model
  (ponderisano brojem primera) i nastavlja trening; pristup je otporan na
  otkaze čvorova, a modeli konvergiraju ka sličnoj tačnosti

**Distribucija treniranja:** horizontalna podela. Svaki čvor dobija svoj
komad od 3.000 letova, namerno mali da bi se korist federacije jasno
videla. Poslednjih 50.000 redova skupa služi kao zajednički test skup.

**Metod evaluacije:** poređenje tačnosti na istom test skupu za četiri
pristupa: uvek pogoditi većinsku klasu, model jednog čvora bez federacije,
centralizovan trening nad svim podacima i federativni model. Merilo uspeha
je da federativni model bude bolji od pojedinačnog i približno jednak
centralizovanom. Probna implementacija daje sledeće vrednosti:

| pristup | podaci | tačnost |
|---|---|---|
| većinska klasa | bez treninga | 0,554 |
| jedan čvor bez federacije | 3.000 | 0,566 |
| centralizovano | 489.000 | 0,566 |
| federativno (5 čvorova) | 5 x 3.000 | 0,573 |

Apsolutna tačnost je ograničena atributima, jer red letenja ne sadrži
stvarne uzroke kašnjenja poput vremenskih prilika.

## Aktori i poruke

| aktor | uloga |
|---|---|
| Membership | HyParView članstvo, po jedan na svakom čvoru |
| Provider | centralni registar članova (provider režim) |
| Member | klijent registra, registruje se i šalje otkucaje |
| Aggregator | FedAvg koordinator: vodi runde, računa ponderisani prosek, meri tačnost |
| Trainer | lokalni trening; u provider režimu odgovara agregatoru, u P2P režimu razmenjuje model sa susedima |

Poruke:

- klaster: Join (zahtev za ulazak), PeerList (spisak poznatih čvorova),
  ForwardJoin (šetnja vesti o novom čvoru, nosi TTL), Ping i Pong (provera
  živosti), NeighborRequest i NeighborResponse (rukovanje za aktivni
  pogled), Disconnect (raskid veze), Register, Heartbeat i Members
  (provider registar), GetPeers i Peers (lokalni upit aktora svom
  membership aktoru: ko su mi susedi)
- federativno učenje: GlobalModel (redni broj runde i globalne težine),
  LocalUpdate (istrenirane težine, broj primera, runda), RegisterTrainer
  (prijava trenera agregatoru), ModelGossip (težine, broj primera i stanje
  G-Countera)
- CRDT sinhronizacija: SyncState (celo stanje brojača)

Skica komunikacije u P2P režimu:

```
 čvor A                         čvor B
+-------------------+          +-------------------+
| Membership <------+--ping---->| Membership       |
|    ^ GetPeers     |          |    ^              |
|    | Peers        |          |    |              |
| Trainer ----------+--Model--->| Trainer          |
|  (trening +       |  Gossip  |  (prosek + merge  |
|   G-Counter)      |<---------+-  + trening)      |
+-------------------+          +-------------------+
```

U provider režimu treneri komuniciraju samo sa agregatorom, u krug:
RegisterTrainer, pa GlobalModel, pa LocalUpdate.

## Detalji implementacije

- Go 1.24, isključivo standardna biblioteka (net, crypto/tls, encoding/gob,
  encoding/csv)
- sertifikati: sopstveni CA, openssl skripta u repozitorijumu
- skup podataka se ne drži u repozitorijumu, preuzima se jednom komandom sa
  OpenML-a
- sistem se demonstrira na više fizičkih mašina; transport je mutual TLS
  preko TCP-a
