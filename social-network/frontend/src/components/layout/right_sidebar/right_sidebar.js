import styles from "./right_sidebar.module.css";
import pfp from "../../../../public/pfp.png";
import Icon from "@/components/shared/icons/Icon";
import useGetUsers from "@/hooks/useGetUsers";
import { useState } from "react";

function FollowCard({ user }) {
  if (!user) return;

  return (
    <div className={styles.element}>
      <div className={styles.details}>
        <img src={pfp.src} alt="pfp" width={50} height={50} />
        <div className={styles.name}>
          <h3>{`${user.firstname} ${user.lastname}`}</h3>
          {user.nickname ? <p>@{user.nickname}</p> : ""}
        </div>
      </div>
      <button className={styles.button}>Follow</button>
    </div>
  );
}

function UsersFollow({ search }) {
  const { users, err, loading } = useGetUsers(search);

  return (
    <div className={styles.sugestions}>
      <h2>Who to follow</h2>
      <div className={styles.cards}>
        {err ? (
          <p>Error Loading the users</p>
        ) : loading ? (
          <p>Loading...</p>
        ) : users.length === 0 ? (
          <p>No users</p>
        ) : (
          users.map((u) => <FollowCard key={u.id} user={u} />)
        )}
      </div>
    </div>
  );
}

export default function RightSideBar() {
  const [search, setSearch] = useState("");

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.search_container}>
          <input
            type="text"
            placeholder="Search"
            value={search}
            onInput={(e) => {
              e.preventDefault();
              setSearch(e.target.value);
            }}
          />
          <Icon name="search" size={24} />
        </div>
      </div>
      <UsersFollow search={search} />
    </div>
  );
}
