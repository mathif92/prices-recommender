import { useState, useEffect, useCallback } from 'react';
import { listHotels, getHotel, triggerCollect } from './api.js';

function Stars({ rating }) {
  const full = Math.round(rating);
  return (
    <span className="inline-flex items-center gap-[1px]" aria-label={`${rating} stars`}>
      {Array.from({ length: 5 }, (_, i) => (
        <span key={i} className={i < full ? 'text-amber-400' : 'text-slate-200'}>
          {i < full ? '\u2605' : '\u2606'}
        </span>
      ))}
      <span className="ml-1 font-semibold text-xs text-slate-500">{rating}</span>
    </span>
  );
}

function fmtDate(d) {
  return d ? d.slice(0, 10) : '';
}

function fmtPrice(price, currency) {
  if (!price || price === 0) return <span className="text-slate-400">Price not available</span>;
  const n = Number.isInteger(price) ? price.toString() : price.toFixed(2);
  return <span className="font-bold text-blue-600">{n} {currency}</span>;
}

function PriceTag({ min, max, currency, count }) {
  if (count === 0) return <span className="text-slate-400 text-sm">No prices</span>;
  if (!min || min === 0) return <span className="text-slate-400 text-sm">Price not available</span>;
  const fmt = (n) => (Number.isInteger(n) ? n.toString() : n.toFixed(2));
  if (min === max) return <span className="font-bold text-blue-600">{fmt(min)} {currency}</span>;
  return <span className="font-bold text-blue-600">{fmt(min)} – {fmt(max)} {currency}</span>;
}

function HotelCard({ hotel, onSelect }) {
  return (
    <article
      className="bg-white border border-slate-200 rounded-xl p-4 shadow-sm hover:shadow-md transition-all hover:-translate-y-0.5 cursor-pointer flex flex-col gap-2"
      onClick={() => onSelect(hotel.id)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && onSelect(hotel.id)}
    >
      <div className="flex items-start justify-between gap-2">
        <h3 className="font-semibold text-base leading-tight">{hotel.name}</h3>
        {hotel.is_all_inclusive && (
          <span className="shrink-0 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700">
            All Inclusive
          </span>
        )}
      </div>
      <div className="flex items-center justify-between gap-2 text-sm">
        <span className="text-slate-500">{hotel.location}</span>
        <Stars rating={hotel.rating} />
      </div>
      {'min_price' in hotel && (
        <div className="flex items-center justify-between pt-2 border-t border-slate-100 text-sm">
          <PriceTag min={hotel.min_price} max={hotel.max_price} currency={hotel.price_currency} count={hotel.price_count} />
          <span className="text-slate-400 text-xs">{hotel.price_count} price{hotel.price_count !== 1 ? 's' : ''}</span>
        </div>
      )}
    </article>
  );
}

function HotelDetail({ id, onBack }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getHotel(id)
      .then(setData)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="flex justify-center py-12"><div className="size-10 border-2 border-slate-200 border-t-blue-600 rounded-full animate-spin" /></div>;
  if (error) return <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm">{error.message}</div>;
  if (!data) return null;

  const { hotel, ratings, reviews, prices } = data;

  return (
    <div>
      <button className="mb-4 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium border border-slate-200 text-slate-700 hover:bg-slate-50 transition cursor-pointer" onClick={onBack}>
        &larr; Back
      </button>
      <div className="mb-6">
        <div className="flex items-start gap-3 flex-wrap">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold mb-1">{hotel.name}</h2>
            <p className="text-slate-500 mb-1">{hotel.location}</p>
            <Stars rating={hotel.rating} />
            {hotel.is_all_inclusive && (
              <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700">
                All Inclusive
              </span>
            )}
          </div>
        </div>
        {hotel.description && <p className="mt-2 text-sm text-slate-500 leading-relaxed">{hotel.description}</p>}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <section className="bg-white border border-slate-200 rounded-xl p-4">
          <h3 className="font-semibold text-sm text-slate-500 uppercase tracking-wide mb-3">Prices ({prices.length})</h3>
          {prices.length === 0 ? (
            <p className="text-center py-8 text-slate-400">No prices recorded</p>
          ) : (
            <div className="divide-y divide-slate-100">
              {prices.map((p, i) => (
                <div key={i} className="flex flex-wrap gap-2 items-center py-2.5 text-sm">
                  {fmtPrice(p.price, p.currency)}
                  <span className="text-slate-400 text-xs">{fmtDate(p.start_date)} – {fmtDate(p.end_date)}</span>
                  {p.checkin_time && <span className="text-slate-400 text-xs ml-auto">Check-in: {p.checkin_time}</span>}
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="bg-white border border-slate-200 rounded-xl p-4">
          <h3 className="font-semibold text-sm text-slate-500 uppercase tracking-wide mb-3">Ratings ({ratings.length})</h3>
          {ratings.length === 0 ? (
            <p className="text-center py-8 text-slate-400">No ratings</p>
          ) : (
            <div className="divide-y divide-slate-100">
              {ratings.map((r, i) => (
                <div key={i} className="flex items-center justify-between py-2.5 text-sm">
                  <Stars rating={r.rating} />
                  <span className="text-slate-400 text-xs">{r.review_count} reviews</span>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="md:col-span-2 bg-white border border-slate-200 rounded-xl p-4">
          <h3 className="font-semibold text-sm text-slate-500 uppercase tracking-wide mb-3">Reviews ({reviews.length})</h3>
          {reviews.length === 0 ? (
            <p className="text-center py-8 text-slate-400">No reviews</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {reviews.map((r, i) => {
                const total = r.positive + r.neutral + r.negative;
                return (
                  <div key={i} className="p-3 rounded-lg bg-slate-50">
                    <div className="flex justify-between items-center mb-2 text-sm">
                      <strong>{r.name}</strong>
                      {r.external_link && (
                        <a href={r.external_link} target="_blank" rel="noopener noreferrer" className="text-xs text-blue-600 hover:underline">
                          Source
                        </a>
                      )}
                    </div>
                    <div className="flex flex-col gap-1.5 text-xs">
                      <div className="grid grid-cols-[4.5rem_1fr_2.5rem] gap-2 items-center">
                        <span className="text-slate-500 text-right">Positive</span>
                        <div className="h-1.5 rounded-full bg-slate-200 overflow-hidden">
                          <div className="h-full rounded-full bg-emerald-500 transition-[width] duration-300" style={{ width: `${(r.positive / total) * 100}%` }} />
                        </div>
                        <span className="font-semibold">{r.positive}</span>
                      </div>
                      <div className="grid grid-cols-[4.5rem_1fr_2.5rem] gap-2 items-center">
                        <span className="text-slate-500 text-right">Neutral</span>
                        <div className="h-1.5 rounded-full bg-slate-200 overflow-hidden">
                          <div className="h-full rounded-full bg-amber-400 transition-[width] duration-300" style={{ width: `${(r.neutral / total) * 100}%` }} />
                        </div>
                        <span className="font-semibold">{r.neutral}</span>
                      </div>
                      <div className="grid grid-cols-[4.5rem_1fr_2.5rem] gap-2 items-center">
                        <span className="text-slate-500 text-right">Negative</span>
                        <div className="h-1.5 rounded-full bg-slate-200 overflow-hidden">
                          <div className="h-full rounded-full bg-red-400 transition-[width] duration-300" style={{ width: `${(r.negative / total) * 100}%` }} />
                        </div>
                        <span className="font-semibold">{r.negative}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}

export default function HotelsPage() {
  const [hotels, setHotels] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [location, setLocation] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [collecting, setCollecting] = useState(false);
  const [collectMsg, setCollectMsg] = useState(null);
  const [selectedId, setSelectedId] = useState(null);
  const limit = 20;
  const totalPages = Math.ceil(total / limit);

  const fetchHotels = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listHotels(page, location);
      setHotels(data.hotels);
      setTotal(data.total);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [page, location]);

  useEffect(() => { fetchHotels(); }, [fetchHotels]);

  const handleSearch = (e) => {
    e.preventDefault();
    setPage(1);
    setLocation(searchInput.trim());
  };

  const handleCollect = async () => {
    setCollecting(true);
    setCollectMsg(null);
    try {
      const data = await triggerCollect();
      setCollectMsg(data.status);
      fetchHotels();
    } catch (e) {
      setCollectMsg(`Error: ${e.message}`);
    } finally {
      setCollecting(false);
    }
  };

  if (selectedId) {
    return <HotelDetail id={selectedId} onBack={() => setSelectedId(null)} />;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight mb-5">Hotels</h1>

      <div className="flex flex-wrap gap-3 items-center mb-5">
        <form onSubmit={handleSearch} className="flex gap-2 flex-1 min-w-0 max-w-md" role="search">
          <input
            className="flex-1 px-3 py-2 rounded-lg border border-slate-200 text-sm bg-white text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 focus:outline-none transition"
            type="search"
            placeholder="Search by location…"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
          <button className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer" type="submit">
            Search
          </button>
          {location && (
            <button
              className="px-4 py-2 rounded-lg text-sm font-medium border border-slate-200 text-slate-700 hover:bg-slate-50 transition cursor-pointer"
              type="button"
              onClick={() => { setLocation(''); setSearchInput(''); setPage(1); }}
            >
              Clear
            </button>
          )}
        </form>
        <button
          className="px-4 py-2 rounded-lg text-sm font-medium bg-emerald-600 text-white hover:bg-emerald-700 transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-1.5 cursor-pointer"
          onClick={handleCollect}
          disabled={collecting}
        >
          {collecting && <span className="size-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />}
          {collecting ? 'Collecting\u2026' : 'Collect Now'}
        </button>
      </div>

      {collectMsg && (
        <div className={`p-3 rounded-lg text-sm mb-4 ${collectMsg.startsWith('Error') ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-700'}`}>
          {collectMsg}
        </div>
      )}

      {error && <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm mb-4">{error}</div>}

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="size-10 border-2 border-slate-200 border-t-blue-600 rounded-full animate-spin" />
        </div>
      ) : hotels.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <p className="mb-1">No hotels found{location && <> for &ldquo;{location}&rdquo;</>}.</p>
          {location && <p className="text-sm">Try a different location name.</p>}
        </div>
      ) : (
        <>
          <p className="text-xs text-slate-500 mb-3">
            {total} hotel{total !== 1 ? 's' : ''}{location && <> for &ldquo;{location}&rdquo;</>}
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {hotels.map((h) => (
              <HotelCard key={h.id} hotel={h} onSelect={setSelectedId} />
            ))}
          </div>

          {totalPages > 1 && (
            <nav className="flex items-center justify-center gap-1.5 mt-5" aria-label="Pagination">
              <button
                className="min-w-9 h-9 inline-flex items-center justify-center rounded-md border border-slate-200 bg-white text-sm cursor-pointer hover:border-blue-500 hover:text-blue-600 transition disabled:opacity-40 disabled:cursor-not-allowed"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                &larr;
              </button>
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  className={`min-w-9 h-9 inline-flex items-center justify-center rounded-md border text-sm font-medium cursor-pointer transition ${
                    p === page
                      ? 'bg-blue-600 border-blue-600 text-white'
                      : 'border-slate-200 bg-white text-slate-700 hover:border-blue-500 hover:text-blue-600'
                  }`}
                  onClick={() => setPage(p)}
                >
                  {p}
                </button>
              ))}
              <button
                className="min-w-9 h-9 inline-flex items-center justify-center rounded-md border border-slate-200 bg-white text-sm cursor-pointer hover:border-blue-500 hover:text-blue-600 transition disabled:opacity-40 disabled:cursor-not-allowed"
                disabled={page >= totalPages}
                onClick={() => setPage(page + 1)}
              >
                &rarr;
              </button>
            </nav>
          )}
        </>
      )}
    </div>
  );
}
