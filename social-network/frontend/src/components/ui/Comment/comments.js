import Image from "next/image";
import pfp from "../../../../public/pfp.png";
import styles from "./comment.module.css";
import { useEffect, useState } from "react";
import fetchClient from "@/lib/api/client";
import useGetComments from "@/hooks/useGetComments";
import Icon from "@/components/shared/icons/Icon";

const dummyComent = {
  id: 1,
  name: "John Doe",
  content: "This is a comment",
};

function Comment({ comment = dummyComent, comments, setComments }) {
  const handleLike = async () => {
    try {
      await fetchClient(`/api/react/comment/${comment.id}`, {
        params: { reaction: comment?.isUserLiked ? "" : "LIKE" },
      });
      const newComments = [...comments];
      const wantedComment = comments.findIndex((c) => c.id === comment.id);
      if (wantedComment !== -1) {
        newComments[wantedComment].isUserLiked =
          !newComments[wantedComment].isUserLiked;
      }
      setComments(newComments);
    } catch (e) {
      console.error(e);
    }
  };
  return (
    <div className={styles.comment_container}>
      <div className={styles.user}>
        <img src={pfp.src} alt="profile picture" width={30} height={30} />
      </div>
      <div className={styles.comment}>
        <span>{comment.name}</span>
        <p>{comment.comment}</p>
        <Icon
          name="heart"
          color="red"
          size={24}
          onClick={handleLike}
          fill={!!comment?.isUserLiked}
          style={{
            cursor: "pointer",
          }}
        />
      </div>
    </div>
  );
}

function CommnetList({ comments, setComments }) {
  return (
    <div className={styles.comments}>
      {comments?.length !== 0 ? (
        comments.map((comment) => (
          <Comment
            key={comment.id}
            comment={comment}
            comments={comments}
            setComments={setComments}
          />
        ))
      ) : (
        <p>No Comment</p>
      )}
    </div>
  );
}

function CommentForm({ post, setComments }) {
  const [comment, setComment] = useState("");
  const [image, setImage] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  async function handleOnSubmit(e) {
    e.preventDefault();
    if (loading) return;
    setError("");
    setLoading(true);
    if (comment.trim().length == 0) {
      setError("Please enter a comment");
      return;
    }
    const newComment = {
      comment: comment.trim(),
      postId: post.id,
    };

    // api call to create a new comment
    try {
      const data = await fetchClient("/api/add/comment", {
        method: "POST",
        body: newComment,
      });
      setComment("");
      setComments((p) => [data.data, ...p]);
    } catch (e) {
      setError(e.message);
      console.log(e);
    } finally {
      setLoading(false);
    }
  }
  function handleFile(e) {
    const file = e.target.files[0];
    if (file) {
      setImage(file);
    }
  }
  return (
    <form onSubmit={handleOnSubmit} className={styles.new_comment}>
      {error != "" && <p>{error}</p>}
      <input
        type="text"
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        placeholder="Write a comment..."
      />
      <input onChange={handleFile} type="file" />
      <input type="submit" />
    </form>
  );
}

export default function Comments({ post, isGroup }) {
  const { comments, setComments, commentError, commentLoading } =
    useGetComments(post, isGroup);
  if (commentLoading) {
    return <p>Loading...</p>;
  }
  if (commentError) {
    return <p>Failed to load comments</p>;
  }
  return (
    <div className={styles.container}>
      <CommnetList comments={comments} setComments={setComments} />
      <CommentForm post={post} setComments={setComments} />
    </div>
  );
}
