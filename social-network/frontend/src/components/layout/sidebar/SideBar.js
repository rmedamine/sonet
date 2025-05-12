"use client";

import styles from "./sidebar.module.css";
import pfp from "../../../../public/pfp.png";
import Image from "next/image";
import Icon from "@/components/shared/icons/Icon";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/providers/AuthProvider";
import ImageElem from "@/components/shared/image/Image";

export default function SideBar() {
  const { logout, user } = useAuth();
  const pathname = usePathname();

  return (
    <div className={styles.sidebar}>
      <h2 className={styles.logo}>Social</h2>
      <div className={styles.sidebar_body}>
        <div className={styles.item_list}>
          <Link
            href="/"
            className={`${styles.item} ${
              pathname === "/" ? styles.active : ""
            }`}
          >
            <Icon name="home" size={20} />
            <span>Home</span>
          </Link>
          <Link
            className={`${styles.item} ${
              pathname === "/profile" ? styles.active : ""
            }`}
            href="/profile"
          >
            <Icon name="user" size={20} />
            <span>Profile</span>
          </Link>
          <Link
            className={`${styles.item} ${
              pathname === "/followers" ? styles.active : ""
            }`}
            href="/followers"
          >
            <Icon name="followers" size={20} />
            <span>Followers</span>
          </Link>
          <Link
            className={`${styles.item} ${
              pathname === "/groups" ? styles.active : ""
            }`}
            href="/groups"
          >
            <Icon name="group" size={20} />
            <span>Groups</span>
          </Link>
          <Link
            className={`${styles.item} ${
              pathname === "/notifications" ? styles.active : ""
            }`}
            href="/notifications"
          >
            <Icon name="notification" size={20} />
            <span>Notifications</span>
          </Link>
          <Link
            className={`${styles.item} ${
              pathname === "/chats" ? styles.active : ""
            }`}
            href="/chats"
          >
            <Icon name="chat" size={20} />
            <span>Chats</span>
          </Link>
        </div>
        <div className={styles.sidebar_fouter}>
          <div className={styles.profile}>
            {user?.avatar ? (
              <ImageElem path={user.avatar} width={50} height={50} />
            ) : (
              <img src={pfp.src} alt="pfp" />
            )}
            <div className={styles.profile_details}>
              <h5>{`${user?.lastname} ${user?.firstname}`}</h5>
              {user?.nickname && <p>@{user.nickname}</p>}
            </div>
          </div>
          <button className={styles.logout} onClick={logout}>
            <Icon name="logout" size={20} />
            <span>Log out</span>
          </button>
        </div>
      </div>
    </div>
  );
}
