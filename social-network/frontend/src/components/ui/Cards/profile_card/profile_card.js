import styles from "./profile_card.module.css";
import pfp from "../../../../../public/pfp.png";
import profile_banner from "../../../../../public/profile_banner.png";
import Image from "next/image";
import fetchClient from "@/lib/api/client";
import { useState, useEffect, useRef } from "react";
import Icon from "@/components/shared/icons/Icon";
import ImageElem from "@/components/shared/image/Image";

export default function ProfileCard({
  profile,
  setProfile,
  ownProfile,
  updateError,
}) {
  console.log(profile);
  const [newProfile, setNewProfile] = useState({ ...profile });
  const [followError, setFollowError] = useState(null);
  const fileInputRef = useRef(null);
  const [editMode, setEditMode] = useState({
    nickname: false,
    email: false,
    dob: false,
    bio: false,
  });
  const [editValues, setEditValues] = useState({
    nickname: profile.nickname || "",
    email: profile.email || "",
    dob: profile.date_of_birth || "",
    bio: profile.about || "",
  });

  useEffect(() => {
    setNewProfile({ ...profile });
    setEditValues({
      nickname: profile.nickname || "",
      email: profile.email || "",
      dob: profile.date_of_birth || "",
      bio: profile.about || "",
    });
  }, [profile]);

  const handleFollow = async (e) => {
    e.preventDefault();
    try {
      await fetchClient(`/api/follow/${profile.id}`, {
        method: "POST",
      });
    } catch (e) {
      setFollowError(e.message);
    }
  };

  const isFollowed = () => {
    const allFollows = profile.followings;
    if (!allFollows || allFollows?.length == 0) return false;
    return allFollows?.find((f) => f.following_id === profile.id);
  };

  const handleEdit = (e, targetName) => {
    e.preventDefault();
    setEditMode({ ...editMode, [targetName]: true });
  };

  const handleChange = (e, field) => {
    setEditValues({ ...editValues, [field]: e.target.value });
  };

  const handlePrivacyToggle = (e) => {
    setNewProfile({ ...newProfile, is_public: !e.target.checked });
  };

  const handleAvatarClick = () => {
    if (ownProfile && fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleAvatarChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      // In a real implementation, you'd upload this file to your server
      // For now, we'll just create a temporary URL
      const imageUrl = URL.createObjectURL(file);
      setNewProfile({ ...newProfile, avatarUrl: imageUrl, avatarFile: file });
    }
  };

  const handleSave = async (field) => {
    try {
      const updatedProfile = { ...newProfile };

      if (field === "nickname") updatedProfile.nickname = editValues.nickname;
      if (field === "email") updatedProfile.email = editValues.email;
      if (field === "dob") updatedProfile.date_of_birth = editValues.dob;
      if (field === "bio") updatedProfile.about = editValues.bio;

      setNewProfile(updatedProfile);
      setEditMode({ ...editMode, [field]: false });
    } catch (error) {
      console.error("Failed to update profile:", error);
    }
  };

  const handleCancel = (field) => {
    // Reset to original value from newProfile, not profile
    setEditValues({
      ...editValues,
      [field]:
        field === "bio"
          ? newProfile.about || ""
          : field === "dob"
          ? newProfile.date_of_birth || ""
          : newProfile[field] || "",
    });
    setEditMode({ ...editMode, [field]: false });
  };

  const saveChanges = async (e) => {
    e.preventDefault();
    setProfile({ ...newProfile });
  };

  const cancelChanges = (e) => {
    e.preventDefault();
    // Reset newProfile to the original profile
    setNewProfile({ ...profile });
    // Reset all edit values
    setEditValues({
      nickname: profile.nickname || "",
      email: profile.email || "",
      dob: profile.date_of_birth || "",
      bio: profile.about || "",
    });
  };

  const renderEditableField = (field, value, placeholder) => {
    if (editMode[field]) {
      return (
        <div className={styles.edit_container}>
          <input
            type={field === "dob" ? "date" : "text"}
            value={editValues[field]}
            onChange={(e) => handleChange(e, field)}
            placeholder={placeholder}
            className={styles.edit_input}
          />
          <div className={styles.edit_actions}>
            <Icon name="check" size={18} onClick={() => handleSave(field)} />
            <Icon name="close" size={18} onClick={() => handleCancel(field)} />
          </div>
        </div>
      );
    }

    return (
      <div className={styles.value}>
        <p>{value}</p>
        {ownProfile && (
          <Icon
            name="edit_pen"
            size={20}
            onClick={(e) => handleEdit(e, field)}
          />
        )}
      </div>
    );
  };

  return (
    <figure className={styles.container}>
      <Image
        src={profile_banner}
        alt="Profile Banner"
        className={styles.banner}
        priority
      />
      <figcaption>
        <div className={styles.main}>
          <div className={styles.pfp_container} onClick={handleAvatarClick}>
            {newProfile.avatar ? (
              <ImageElem
                src={newProfile.avatar}
                className={styles.pfp}
                alt={"pfp"}
              />
            ) : (
              <img src={pfp.src} alt="pfp" className={styles.pfp} />
            )}
            {ownProfile && (
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleAvatarChange}
                accept="image/*"
                className={styles.file_input}
              />
            )}
          </div>
          <div className={styles.name}>
            <h2
              className={styles.name}
            >{`${newProfile.firstname} ${newProfile.lastname}`}</h2>
            {renderEditableField(
              "nickname",
              newProfile?.nickname && `@${newProfile.nickname}`,
              "Enter nickname"
            )}
          </div>
          <div className={styles.stats}>
            <div className={styles.stat}>
              <strong>{newProfile?.followers?.length || 0}</strong>
              <p>Followers</p>
            </div>
            <div className={styles.stat}>
              <strong>{newProfile?.following?.length || 0}</strong>
              <p>Following</p>
            </div>
          </div>
        </div>
        <div className={styles.details}>
          {ownProfile && (
            <div className={styles.item}>
              <strong>Private mode</strong>
              <label className={styles.switch}>
                <input
                  type="checkbox"
                  checked={!newProfile.is_public}
                  onChange={handlePrivacyToggle}
                />
                <span className={styles.slider}></span>
              </label>
            </div>
          )}
          <div className={styles.item}>
            <strong>Email:</strong>
            {renderEditableField("email", newProfile.email, "Enter email")}
          </div>
          <div className={styles.item}>
            <strong>Date of birth:</strong>
            {renderEditableField(
              "dob",
              newProfile.date_of_birth,
              "Enter date of birth"
            )}
          </div>
          <div className={`${styles.item} ${styles.bio}`}>
            <strong>About me: </strong>
            {renderEditableField(
              "bio",
              newProfile.about || "No bio",
              "Tell us about yourself"
            )}
          </div>
          {(!ownProfile || isFollowed()) && (
            <div className={styles.actions}>
              <button onClick={handleFollow}>Follow</button>
              {followError && <p>{followError}</p>}
            </div>
          )}
        </div>
      </figcaption>
      {ownProfile && (
        <div className={styles.actions}>
          <button className={styles.save_button} onClick={saveChanges}>
            Save
          </button>
          <button className={styles.cancel_button} onClick={cancelChanges}>
            Cancel
          </button>
        </div>
      )}
      {updateError && <p>{updateError}</p>}
    </figure>
  );
}
