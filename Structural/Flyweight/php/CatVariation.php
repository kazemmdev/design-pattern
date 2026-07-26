<?php

namespace DesignPatterns\Structural\Flyweight;

/**
 * Flyweight objects represent the data shared by multiple Cat objects. This is
 * the combination of breed, color, texture, etc.
 */
class CatVariation
{
    /**
     * The so-called "intrinsic" state.
     */
    public $breed;

    public $image;

    public $color;

    public $texture;

    public $fur;

    public $size;

    public function __construct(
        string $breed,
        string $image,
        string $color,
        string $texture,
        string $fur,
        string $size
    )
    {
        $this->breed = $breed;
        $this->image = $image;
        $this->color = $color;
        $this->texture = $texture;
        $this->fur = $fur;
        $this->size = $size;
    }

    /**
     * This method displays the cat information. The method accepts the
     * extrinsic  state as arguments. The rest of the state is stored inside
     * Flyweight's fields.
     *
     * You might be wondering why we had put the primary cat's logic into the
     * CatVariation class instead of keeping it in the Cat class. I agree, it
     * does sound confusing.
     *
     * Keep in mind that in the real world, the Flyweight pattern can either be
     * implemented from the start or forced onto an existing application
     * whenever the developers realize they've hit upon a RAM problem.
     *
     * In the latter case, you end up with such classes as we have here. We kind
     * of "refactored" an ideal app where all the data was initially inside the
     * Cat class. If we had implemented the Flyweight from the start, our class
     * names might be different and less confusing. For example, Cat and
     * CatContext.
     *
     * However, the actual reason why the primary behavior should live in the
     * Flyweight class is that you might not have the Context class declared at
     * all. The context data might be stored in an array or some other more
     * efficient data structure. You won't have another place to put your
     * methods in, except the Flyweight class.
     *
     * @param string $name
     * @param string $age
     * @param string $owner
     * @return array
     */
    public function renderProfile(string $name, string $age, string $owner): array
    {
        return [
            "Name" => $name,
            "Age" => $age,
            "Owner" => $owner,
            "Breed" => $this->breed,
            "Image" => $this->image,
            "Color" => $this->color,
            "Texture" => $this->texture,
        ];
    }
}