import { useState, useEffect } from 'react';
import { ListingCard } from './ListingCard';
import { Button } from './ui/button';
import { User, Moon, Sun } from 'lucide-react';

interface Listing {
  id: number;
  image: string;
  title: string;
  description: string;
  username: string;
  isPremium: boolean;
  category: string;
}

interface ProfileTabProps {
  isDark: boolean;
  toggleTheme: () => void;
}

export function ProfileTab({ isDark, toggleTheme }: ProfileTabProps) {
  const [listings, setListings] = useState<Listing[]>([]);
  const [loading, setLoading] = useState(true);
  const [userId] = useState<string | null>(() => {
    // Get user_id from URL params or localStorage
    const params = new URLSearchParams(window.location.search);
    return params.get('user_id') || localStorage.getItem('user_id');
  });

  useEffect(() => {
    if (userId) {
      fetchMyAds();
    } else {
      setLoading(false);
    }
  }, [userId]);

  const fetchMyAds = async () => {
    if (!userId) return;
    
    setLoading(true);
    try {
      const response = await fetch(`/api/myads?user_id=${userId}`);
      const data = await response.json();
      
      const transformedListings = data.map((ad: any) => ({
        id: ad.id,
        image: ad.photo_id || 'https://via.placeholder.com/800x450?text=No+Image',
        title: ad.title,
        description: ad.desc,
        username: `@${ad.username}`,
        isPremium: ad.is_premium,
        category: ad.category,
      }));
      
      setListings(transformedListings);
    } catch (error) {
      console.error('Failed to fetch my ads:', error);
      setListings([]);
    } finally {
      setLoading(false);
    }
  };

  const hasListings = listings.length > 0;

  return (
    <div className="p-4 space-y-6">
      {/* Header with Theme Toggle */}
      <div className="pt-2">
        <div className="flex items-center justify-between mb-2">
          <h1 className="text-2xl">Профиль</h1>
          <Button
            onClick={toggleTheme}
            variant="outline"
            size="icon"
            className="rounded-xl border-border"
          >
            {isDark ? <Sun size={20} /> : <Moon size={20} />}
          </Button>
        </div>
        <p className="text-muted-foreground">Управление вашими объявлениями</p>
      </div>

      {/* No Listings State */}
      {!hasListings && (
        <div className="flex flex-col items-center justify-center py-16 space-y-6">
          <div className="w-24 h-24 bg-muted rounded-full flex items-center justify-center">
            <User size={48} className="text-muted-foreground" />
          </div>
          
          <div className="text-center space-y-2">
            <p className="text-muted-foreground">У вас нет объявлений</p>
            <p className="text-muted-foreground/60 text-sm">Создайте своё первое объявление</p>
          </div>

          <Button
            className="bg-[#FF0000] hover:bg-[#CC0000] text-white px-8 py-6 rounded-2xl shadow-lg"
          >
            Обратитесь к менеджеру
          </Button>
        </div>
      )}

      {/* Listings State */}
      {loading ? (
        <div className="text-center py-12 text-muted-foreground">
          <p>Загрузка объявлений...</p>
        </div>
      ) : hasListings ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-muted-foreground">Мои объявления</p>
            <Button
              className="bg-[#FF0000] hover:bg-[#CC0000] text-white rounded-xl"
              size="sm"
            >
              Создать
            </Button>
          </div>
          
          {listings.map((listing) => (
            <ListingCard key={listing.id} listing={listing} />
          ))}
        </div>
      ) : null}
    </div>
  );
}
