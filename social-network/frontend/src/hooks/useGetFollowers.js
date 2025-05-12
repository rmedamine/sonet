import fetchClient from "@/lib/api/client";
import { useEffect, useState } from "react";

export default function useGetFollowers(userId, activeTab) {
  const [followers, setFollowers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const getFollowers = async () => {
    if (!userId) {
      setError("User ID is required");
      return;
    }
    setError(null);
    try {
      const data = await fetchClient(
        `/api/${
          activeTab === "followings" ? "followings" : "followers"
        }/${userId}`
      );
      setFollowers(data.data || []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    getFollowers();
  }, []);

  return {
    followers,
    followersLoading: loading,
    followersError: error,
  };
}
